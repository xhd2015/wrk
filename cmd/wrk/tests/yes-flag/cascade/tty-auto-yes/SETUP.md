# Scenario

**Feature**: TTY `wrk --done -y` auto-confirms cascaded ahead external dep merge-back

```
consumer wt + ahead external dep (no local replace) -> TTY wrk --done -y -> both wts removed, dep merged
```

## Steps

1. Create consumer wt and external dep wt via `wrk --bring`.
2. Commit ahead on external dep wt.
3. Drop consumer `replace => ./external/...` so consumer `--done` can finish after cascade.
4. Run `wrk --done -y` under fake TTY.

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const yesCascadeDepModule = "example.com/dep"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runYesCascadeGoMod(t, mainRepo, "edit", "-require="+yesCascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+yesCascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--bring", depRepo)
	req.ExternalWtDir = externalPath

	writeFile(t, filepath.Join(externalPath, "dep.go"), "package dep // tty fix\n")
	runGitIsolated(t, externalPath, "add", "dep.go")
	runGitIsolated(t, externalPath, "commit", "-m", "dep fix on external worktree")

	// Remove replace so consumer merge-back succeeds after cascade.
	runYesCascadeGoMod(t, wtDir, "edit", "-dropreplace="+yesCascadeDepModule)
	runYesCascadeGoMod(t, wtDir, "edit", "-droprequire="+yesCascadeDepModule)
	// go mod edit can leave go.sum / external/ dirty; --done refuses uncommitted changes.
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "drop dep replace for done", "--allow-empty")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "-y"}
	req.UseScriptTTY = true
	return nil
}

func runYesCascadeGoMod(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```
