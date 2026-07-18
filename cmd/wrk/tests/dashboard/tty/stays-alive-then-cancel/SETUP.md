# Scenario

**Feature**: TTY bare `wrk` stays alive with dashboard UI; `q` cancels without compose

```
linked-wt (ahead) -> tty-watch detach wrk
  -> snapshot has dashboard BEFORE exit (process still alive)
  -> send q
  -> [Terminal exited]
  -> no new worktrees; compose argv log empty
```

## Steps

1. Setup linked worktree ahead of main.
2. Start bare `wrk` under tty-watch (`TERM=dumb`, isolated WRK_HOME).
3. Assert live snapshot looks like dashboard (not already exited).
4. Send `q`; wait for terminal exit.
5. Assert no compose / no create.

```go
func Setup(t *testing.T, req *Request) error {
	setupDashboardLinkedAhead(t, req)
	// Root Run may execute non-TTY bare wrk (harmless); ASSERT drives tty-watch.
	req.Args = nil
	return nil
}
```
