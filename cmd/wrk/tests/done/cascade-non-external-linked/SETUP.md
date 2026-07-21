# Scenario

**Feature**: wrk --done cascades merge-back to linked worktrees outside external/

```
# consumer wt + manual linked wt at deps/foo (not under external/) -> wrk --done
# scan_repo.Scan discovers deps/foo; cascade removes it, then consumer merge-back
consumer wt + deps/foo linked wt -> wrk --done -> deps/foo removed, consumer exit 0
```

## Steps

1. Create consumer main repo with a clean `go.mod` (no local replace to `deps/foo`).
2. `wrk` creates the consumer linked worktree.
3. Create a dep main repo and run `git worktree add` into `{consumerWt}/deps/foo`.
4. Run `wrk --done` from the consumer worktree.

```go
import (
	"os/exec"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "foodep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module example.com/foodep\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "foo.go"), "package foo\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "foo.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	depsWtDir := filepath.Join(wtDir, "deps", "foo")
	req.DepsLinkedWtDir = depsWtDir
	depBranch := branchName("main", wrkDate, 0)
	runGitIsolated(t, depRepo, "worktree", "add", "-b", depBranch, depsWtDir)

	// Keep consumer wt clean so --done is not blocked by untracked deps/ files.
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/deps\n")
	runGitIsolated(t, wtDir, "add", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "ignore deps worktrees")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```