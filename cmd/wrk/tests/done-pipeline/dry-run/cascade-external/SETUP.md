# Scenario

**Feature**: cascade dry-run plans nested linked worktree merge-back without removing it

```
# consumer wt + external/* dep wt (already-included; replace dropped so guard does not block)
consumer + external/mydep-…  -> wrk --done --dry-run
  -> would: cascade merge-back <external-path>  (compact plan; no real MergeBack)
  -> primary dry-run plan for consumer
  -> external still on disk after exit 0
```

## Steps

1. Consumer main with go.mod requiring `example.com/dep`.
2. `wrk` → consumer linked worktree; `wrk --dep` → external dep worktree.
3. Drop consumer replace so local-replace guard does not block primary planning.
4. Commit clean ignore for `external/`.
5. Snapshot baseline; run `wrk --done --dry-run` (no `-y`).

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const dryRunCascadeDepModule = "example.com/dep"

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
	runGoModDryRunCascade(t, mainRepo, "edit", "-require="+dryRunCascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+dryRunCascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	// Drop replace so primary dry-run is not blocked by local-replace guard after cascade plan.
	runGoModDryRunCascade(t, wtDir, "edit", "-dropreplace="+dryRunCascadeDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	// Consumer is clean + already-included relative to main → dry-run plans remove-only
	// for consumer; cascade must still plan external without removing it.
	req.RepoDir = wtDir
	recordComposeDryRunBaseline(t, req)
	req.Args = []string{"--done", "--dry-run"}
	return nil
}

func runGoModDryRunCascade(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```
