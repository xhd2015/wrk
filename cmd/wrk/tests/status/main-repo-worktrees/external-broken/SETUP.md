# Scenario

**Feature**: appended minimal block for broken external worktree

```
# external wt checkout exists but git metadata is broken (stale gitdir)
external wt (alive, git fails) -> appended Dir + Status: error: ...
```

## Steps

1. Create external wrk worktree from main repo.
2. Overwrite external wt `.git` with stale `gitdir:` path.
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, _ := createExternalWrkWorktree(t, req)
	breakWorktreeGitMetadata(t, req, wtDir)
	req.RepoDir = mainRepo
	return nil
}
```