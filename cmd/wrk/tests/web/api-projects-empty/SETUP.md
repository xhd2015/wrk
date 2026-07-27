# Scenario

**Feature**: wrk --web exposes GET /api/wrk/projects with empty projects array

```
# empty WRK_HOME (no projects.json entries)
wrk --web --port <free>
  -> HTTP GET /api/wrk/projects → 200
  -> JSON {"projects":[]}  (array never null)
  -> process killed after probe
```

## Steps

1. Pick a free localhost port; leave `WrkHome` empty (root Setup mkdir only).
2. Enable `WebProbe` for path `/api/wrk/projects`.
3. Root `Run` starts wrk, GETs the API path, then kills the server.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// WRK_HOME is empty (no projects.json) — ListProjects returns [].
	setupWebProbe(t, req, "/api/wrk/projects")
	return nil
}
```
