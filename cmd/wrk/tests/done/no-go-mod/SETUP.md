# Scenario

**Feature**: wrk --done merges back a linked worktree whose checkout has no go.mod

```
# repo has NO go.mod; wrk creates a linked worktree; --done should merge-back,
# not error with "no go.mod found within ..."
myrepo (main, no go.mod) + wt (main-{date}) -> wrk --done -> remove only
```

## Steps

1. Create main repo on `main` with a README commit but NO `go.mod`.
2. Create a linked worktree via `wrk` (create does not require go.mod).
3. Run `wrk --done` from the linked worktree root (already-included → remove only).

## Expected (correct) behavior

The go.mod lookup in `runDone` exists only to feed the local-filesystem-replace
guard. A checkout with no go.mod cannot host a local replace, so the guard is
trivially satisfied and `--done` must proceed with the merge-back. The worktree
is at the same commit as main (already-included), so `--done` removes the
worktree + branch and exits 0.

## Bug

Current behavior: `findGoModDir` hard-errors `no go.mod found within <top>`
before merge-back runs, so exit is non-zero and the worktree is left in place.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	// Deliberately NO go.mod: reproduces "no go.mod found within ..." from
	// `wrk --done` in a non-Go checkout (e.g. the dot-pkgs-ai report).

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	branch := branchName("main", wrkDate, 0)
	req.WtBranch = branch

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
