# Scenario

**Feature**: two out-of-tree main-owned worktrees keep ListLinked primary order (no header)

```
# wrk twice from same main repo → two out-of-tree wts
myrepo -> wrk -> wt1; wrk -> wt2

# primary order follows ListLinked porcelain; external empty → no header
wrk --status from main -> primary: main, wt1, wt2
```

## Steps

1. Create first external wrk worktree from main repo.
2. Run `wrk` again from the same main repo (collision suffix `-1`).
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	createSecondExternalWrkWorktree(t, req, mainRepo)
	req.RepoDir = mainRepo
	return nil
}
```