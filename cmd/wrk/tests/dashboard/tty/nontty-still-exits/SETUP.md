# Scenario

**Feature**: regression — non-TTY bare `wrk` still one-shot static snapshot (exits)

```
linked-wt (clean) -> wrk  (harness pipes; no ACTION)
  -> exit 0 immediately
  -> fine-grained static dashboard snapshot
```

## Steps

1. Clean linked worktree.
2. Run bare `wrk` via normal harness (non-TTY).
3. Assert P2 snapshot core (process does not hang).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardLinkedWorktree(t, req)
	req.Args = nil
	req.ExtraEnv = nil
	return nil
}
```
