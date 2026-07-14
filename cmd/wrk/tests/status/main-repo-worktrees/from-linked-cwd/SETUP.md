# Scenario

**Feature**: --status from external linked worktree cwd skips append phase

```
# main repo has external wt; cwd is inside that external wt
external wt cwd -> wrk --status -> scan only, no appended section
```

## Steps

1. Create external wrk worktree from main repo.
2. Run `wrk --status` with cwd set to the external worktree root.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, branch := createExternalWrkWorktree(t, req)
	req.RepoDir = wtDir
	req.WtBranch = branch
	return nil
}
```