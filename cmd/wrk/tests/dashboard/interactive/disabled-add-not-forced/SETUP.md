# Scenario

**Feature**: disabled **Add changes** gate cannot inject `--add-all` even if toggles force on

```
linked-wt (clean porcelain, Add changes [-])
  + WRK_DASHBOARD_ACTION=run-done
  + WRK_DASHBOARD_DRY_RUN=1
  + WRK_DASHBOARD_TOGGLES=add-changes=on
  -> compose argv must NOT include --add-all
  -> exit 0 (force-on ignored) or non-zero clear gate error — either OK if no --add-all applied
```

## Steps

1. Clean linked worktree (no untracked/unstaged).
2. Force action run-done dry-run with toggle `add-changes=on`.
3. Assert recipe omits `--add-all`.

```go
func Setup(t *testing.T, req *Request) error {
	// Clean linked (no dirty) so Add changes is gated [-].
	setupDashboardLinkedWorktree(t, req)
	setDashboardAction(t, req, "run-done", true /* dryRun */)
	setDashboardToggles(req, "add-changes=on")
	req.Args = nil
	return nil
}
```
