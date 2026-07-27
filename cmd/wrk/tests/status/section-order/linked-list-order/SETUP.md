# Scenario

**Feature**: two main-owned out-of-tree linked worktrees keep ListLinked order; no header

```
# wrk twice -> two WRK external wts
myrepo -> wt1; wrk again -> wt2
wrk --status -> primary main + ListLinked order; no ---- external ----
```

## Steps

1. Create first WRK external worktree from main.
2. Create second WRK external worktree (collision suffix).
3. Run `wrk --status` from main root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, _, _ := setupWrkWorktreeFromMain(t, req)
	createSecondExternalWrkWorktree(t, req, mainRepo)
	req.RepoDir = mainRepo
	return nil
}
```
