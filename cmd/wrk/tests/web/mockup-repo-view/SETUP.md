# Scenario

**Feature**: wrk --web SPA serves client route `/mockup/repo-view`

```
# free port + empty WRK_HOME
wrk --web --port <free>
  -> HTTP GET /mockup/repo-view → 200 text/html (SPA shell)
  -> body is HTML app shell (id="root" / wrk markers)
  -> process killed after probe
```

## Steps

1. Pick a free localhost port.
2. Enable `WebProbe` for path `/mockup/repo-view`.
3. Root `Run` starts wrk, waits for listen URL, GETs the path, then SIGTERM/SIGKILL.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupWebProbe(t, req, "/mockup/repo-view")
	return nil
}
```
