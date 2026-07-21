# Scenario

**Feature**: wrk --done from a main repo without go.mod reports "not a linked worktree"

```
# cwd is main repo checkout (no go.mod, not a linked worktree)
myrepo (main, no go.mod) -> wrk --done -> not a linked worktree
```

## Steps

1. Create main repo on `main` with a README commit but NO `go.mod`.
2. Run `wrk --done` from the main repo root.

## Expected (correct) behavior

The go.mod lookup must not gate the linked-worktree check. A main-repo checkout
without go.mod is still a main repo (not a linked worktree), so `--done` should
fail with `not a linked worktree` — the same diagnostic `done/not-linked` asserts
for the go.mod-having case. The absence of a go.mod is irrelevant: no local
replace can exist, so the guard is a no-op and the linked-worktree check decides.

## Bug

Current behavior: `findGoModDir` errors `no go.mod found within <top>` before
the linked-worktree check runs, masking the real `not a linked worktree`
diagnostic. The sibling `done/not-linked` test works around this by writing a
go.mod; this leaf drops the go.mod to expose the masking.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	// Deliberately NO go.mod: the linked-worktree check must still run.

	req.RepoDir = mainRepo
	req.Args = []string{"--done"}
	return nil
}
```
