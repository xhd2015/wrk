# Scenario

**Feature**: wrk --done cascades merge-back to external/* worktrees before local-replace guard

```
# consumer wt + external/dep wt (from --bring) -> wrk --done -> cascade removes dep wt, parent errors on local replace
consumer wt -> wrk --bring -> external wt + replace in go.mod -> wrk --done -> dep wt gone, parent non-zero
```

## Steps

1. Create consumer main repo with go.mod requiring `example.com/dep`.
2. Run `wrk` to create consumer linked worktree.
3. Run `wrk --bring` from consumer wt to spawn `external/mydep-main-{date}`.
4. Run `wrk --done` from consumer wt (replace still present).

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const cascadeDepModule = "example.com/dep"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoMod(t, mainRepo, "edit", "-require="+cascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--bring", depRepo)
	req.ExternalWtDir = externalPath

	// Commit --bring replace/gitignore so D2 dirty preflight does not block
	// cascade; replace remains in go.mod for the local-replace guard under test.
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "commit dep replace and ignore", "--allow-empty")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}

func runGoMod(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```