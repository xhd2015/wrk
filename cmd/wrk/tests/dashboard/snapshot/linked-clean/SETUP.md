# Scenario

**Feature**: non-TTY bare `wrk` from a clean linked worktree disables add changes with `[-]`

```
linked-wt (clean porcelain)
  -> wrk (non-TTY, no args)
  -> add changes row [-] (disabled; no unstaged/untracked)
  -> MERGE BACK default selected; no create-hint
```

## Steps

1. Clean linked worktree.
2. Bare `wrk` non-TTY.

```go
func Setup(t *testing.T, req *Request) error {
	linked := setupDashboardLinkedWorktree(t, req)
	req.RepoDir = linked
	req.Args = nil
	return nil
}
```
