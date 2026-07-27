# Scenario

**Feature**: dry-run with nested external cascade shows phase headers and cascade plan lines under cascade phase

```
# consumer + external/* dep wt (already-included; replace dropped so primary can plan)
consumer + external/mydep-…
  -> wrk --done --dry-run
  -> ==> cascade
  -> would: cascade merge-back <external-path>
  -> ==> own
  -> primary MergeBack dry-run plan
  -> external + consumer still on disk
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

const doneOutputCascadeDepModule = "example.com/dep"

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
	runGoModDryRunCascadeUX(t, mainRepo, "edit", "-require="+doneOutputCascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+doneOutputCascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	// Drop replace so primary dry-run is not blocked by local-replace guard after cascade plan.
	runGoModDryRunCascadeUX(t, wtDir, "edit", "-dropreplace="+doneOutputCascadeDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--dry-run"}
	return nil
}

func runGoModDryRunCascadeUX(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```
