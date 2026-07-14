# Scenario

**Feature**: wrk --done --no-in-module-replace blocks even an intra-repo filesystem replace

```
# same intra-repo replace => ./submod, but --no-in-module-replace restores strict guard
consumer wt (replace => ./submod, same repo) -> wrk --done --no-in-module-replace -> block
```

## Steps

1. Same setup as `intra-replace-warns`: consumer `go.mod` with
   `replace example.com/foo => ./submod` pointing at a real nested module in the
   same repo (intra-repo).
2. Create a linked worktree via `wrk`.
3. Run `wrk --done --no-in-module-replace` from the linked worktree.

## Expected (correct) behavior

With `--no-in-module-replace`, the guard is fully strict: **every** local
filesystem replace blocks, including intra-repo ones. So the `./submod` replace
blocks `--done` (non-zero exit), merge-back does not run, and the worktree
remains.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)

	mkdirAll(t, filepath.Join(mainRepo, "submod"))
	writeFile(t, filepath.Join(mainRepo, "submod", "go.mod"), "module example.com/foo\n\ngo 1.21\n")
	writeFile(t, filepath.Join(mainRepo, "go.mod"),
		"module example.com/consumer\n\ngo 1.22\n\nreplace example.com/foo => ./submod\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "submod")
	// --no-verify bypasses the project's go-no-local-replace pre-commit hook
	// (same rationale as intra-replace-warns).
	runGitIsolated(t, mainRepo, "commit", "--no-verify", "-m", "add consumer with intra-repo replace")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--no-in-module-replace"}
	return nil
}
```
