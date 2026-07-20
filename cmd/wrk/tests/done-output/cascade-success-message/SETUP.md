# Scenario

**Feature**: real cascade success always prints each target’s MergeBack **Message** on stdout (D5)

```
# clean contained external + clean own (replace dropped)
  -> wrk --done
  -> ==> cascade
  -> worktree removed: <external> …   (cascade Message; remove-only when contained)
  -> ==> own
  -> merged branch … / worktree removed: <consumer> …
  -> external + consumer removed; exit 0
```

## Preconditions

- Classic TDD: cascade success is silent today (no Message print in
  `mergeBackExternalWorktree`) → leaf **RED** until D5.
- Inherits `done-output/` phase helpers (`assertDonePhaseHeaders`, `assertNoANSIEscape`).

## Steps

1. Consumer main + linked wt + `--dep` external (contained / already-included).
2. Drop replace; commit gitignore so own is clean enough to finish.
3. Run bare `wrk --done`.

```go
import (
	"os/exec"
	"path/filepath"
)

const doneOutputSuccessCascadeDepModule = "example.com/dep-msg"

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, mainRepo, "edit", "-require="+doneOutputSuccessCascadeDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+doneOutputSuccessCascadeDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	runGoModInDir(t, wtDir, "edit", "-dropreplace="+doneOutputSuccessCascadeDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
