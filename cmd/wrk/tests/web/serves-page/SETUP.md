# Scenario

**Feature**: wrk --web serves the standalone HTML workflow page at GET /

```
# free port + empty WRK_HOME
wrk --web --port <free>
  -> stdout: http://127.0.0.1:<port>/\n
  -> HTTP GET / → 200 text/html
  -> body includes workflow markers (task, Main, Remote, worktree/changes, wrk)
  -> process killed after probe (never hangs suite)
```

## Steps

1. Pick a free localhost port.
2. Enable `WebProbe` for path `/`.
3. Root `Run` starts wrk, waits for listen URL, GETs `/`, then SIGTERM/SIGKILL.

```go
func Setup(t *testing.T, req *Request) error {
	setupWebProbe(t, req, "/")
	return nil
}
```
