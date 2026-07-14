# Scenario

**Feature**: appended minimal block for prunable external worktree

```
# external wt registered in git but checkout directory removed
external wt (prunable) -> appended Dir + Status: prunable
```

## Steps

1. Create external wrk worktree from main repo.
2. Delete the external checkout directory (leave git registration).
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, _ := createExternalWrkWorktree(t, req)
	removeWorktreeCheckout(t, wtDir)
	req.RepoDir = mainRepo
	return nil
}
```