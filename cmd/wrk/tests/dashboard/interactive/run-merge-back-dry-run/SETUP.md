# Scenario

**Feature**: `WRK_DASHBOARD_ACTION=run-merge-back` + dry-run runs real MERGE BACK compose

```
linked-wt (ahead) + ACTION=run-merge-back + DRY_RUN=1
  -> bare wrk
  -> compose argv uses --merge-back (not --done)
  -> dry-run plan without worktree remove
  -> exit 0; worktree stays
```

## Steps

1. Linked worktree with commit ahead.
2. `WRK_DASHBOARD_ACTION=run-merge-back`, `WRK_DASHBOARD_DRY_RUN=1`, argv log.
3. Run bare `wrk` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardLinkedAhead(t, req)
	setDashboardAction(t, req, "run-merge-back", true /* dryRun */)
	req.Args = nil
	return nil
}
```
