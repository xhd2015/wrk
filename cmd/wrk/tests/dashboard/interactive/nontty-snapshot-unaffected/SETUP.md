# Scenario

**Feature**: non-TTY bare `wrk` without `WRK_DASHBOARD_ACTION` stays P2 static snapshot

```
linked-wt (clean) -> wrk  (no ACTION env)
  -> exit 0
  -> full fine-grained dashboard snapshot (glyphs, stages, no create-hint)
  -> compose argv log unused / empty
```

## Steps

1. Clean linked worktree.
2. Do **not** set `WRK_DASHBOARD_ACTION`.
3. Run bare `wrk`; assert P2 snapshot core.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardLinkedWorktree(t, req)
	// Explicit: no interactive action env (P2 path).
	req.ExtraEnv = nil
	req.Args = nil
	return nil
}
```
