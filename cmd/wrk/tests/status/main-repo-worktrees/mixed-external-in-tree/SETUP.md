# Scenario

**Feature**: scan shows in-tree linked wt; append shows external only

```
# wrk external + in-tree git worktree add
myrepo -> wrk external + wt-linked in-tree

# scan: `.` + `wt-linked`; append: external abs block only
wrk --status from main -> dedup in-tree, append external
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