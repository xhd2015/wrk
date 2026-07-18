# Scenario

**Feature**: non-TTY bare `wrk` from a dirty linked worktree shows full snapshot with add changes enabled

```
linked-wt (dirty untracked)
  -> wrk (non-TTY, no args)
  -> exit 0; dashboard snapshot
  -> add changes row enabled ([•] or [ ])
  -> Pre/Main/After + MERGE BACK default [•] + Batch would-run; no create-hint
```

## Steps

1. Linked worktree with untracked dirty file.
2. Bare `wrk` non-TTY from linked cwd.

```go
func Setup(t *testing.T, req *Request) error {
	linked := setupDashboardLinkedWorktree(t, req)
	markDashboardDirtyUntracked(t, linked)
	req.RepoDir = linked
	req.Args = nil
	return nil
}
```
