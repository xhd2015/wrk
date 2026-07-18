# Scenario

**Feature**: `WRK_DASHBOARD_ACTION=run-done` + dry-run runs real DONE multi-stage compose

```
linked-wt (ahead, dirty untracked) + ACTION=run-done + DRY_RUN=1
  -> bare wrk
  -> compose argv log = default DONE recipe + --dry-run + --add-all
  -> real dry-run plan evidence (not static snapshot only)
  -> exit 0; worktree still present (dry-run)
```

## Steps

1. Linked worktree ahead of main.
2. Mark untracked dirty so **Add changes** defaults on → `--add-all`.
3. `WRK_DASHBOARD_ACTION=run-done`, `WRK_DASHBOARD_DRY_RUN=1`, argv log path.
4. Run bare `wrk` from linked worktree.

```go
func Setup(t *testing.T, req *Request) error {
	linked := setupDashboardLinkedAhead(t, req)
	markDashboardDirtyUntracked(t, linked)
	setDashboardAction(t, req, "run-done", true /* dryRun */)
	req.Args = nil
	return nil
}
```
