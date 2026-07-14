# Scenario

**Feature**: two external worktrees appended in git worktree list order

```
# wrk twice from same main repo → two external wts
myrepo -> wrk -> wt1; wrk -> wt2

# append order follows ListLinked porcelain order
wrk --status from main -> scan `.` + append wt1 then wt2
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