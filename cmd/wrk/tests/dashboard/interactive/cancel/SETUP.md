# Scenario

**Feature**: `WRK_DASHBOARD_ACTION=cancel` exits 0 without compose mutations

```
linked-wt (ahead) + WRK_DASHBOARD_ACTION=cancel
  -> bare wrk
  -> exit 0
  -> worktree still linked; branch kept
  -> events.jsonl command=dashboard
  -> compose argv log stays empty (no RUN)
```

## Steps

1. Create main + linked worktree with a commit ahead.
2. Set `WRK_DASHBOARD_ACTION=cancel` (and argv log path).
3. Run bare `wrk` from linked worktree (non-TTY harness).

```go
func Setup(t *testing.T, req *Request) error {
	setupDashboardLinkedAhead(t, req)
	setDashboardAction(t, req, "cancel", false /* dryRun */)
	req.Args = nil
	return nil
}
```
