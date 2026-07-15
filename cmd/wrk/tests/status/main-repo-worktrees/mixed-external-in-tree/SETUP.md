# Scenario

**Feature**: main-repo status primary order follows ListLinked for in-tree + out-of-tree linked

```
# wrk out-of-tree + in-tree git worktree add (both main-owned → primary)
myrepo -> wrk external + wt-linked in-tree

# primary: main then ListLinked porcelain order; external empty → no header
wrk --status from main -> primary three blocks; no "---- external ----"
```

## Steps

1. Create external wrk worktree from main repo.
2. Add in-tree linked worktree at `myrepo/wt-linked`.
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	wtDir := addInTreeLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")

	req.RepoDir = mainRepo
	req.InTreeWtDir = wtDir
	req.InTreeWtRel = "wt-linked"
	req.InTreeWtBranch = "wt-side"
	return nil
}
```
