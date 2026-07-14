# Scenario

**Feature**: appended external worktree shows dirty status counts

```
# external wt with uncommitted tracked change
external wt (dirty) -> appended Status: dirty (...)
```

## Steps

1. Create external wrk worktree from main repo.
2. Modify `README.md` in the external worktree without committing.
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, _ := createExternalWrkWorktree(t, req)
	dirtyWorktreeFile(t, wtDir, "README.md", "# dirty external\n")
	req.RepoDir = mainRepo
	return nil
}
```