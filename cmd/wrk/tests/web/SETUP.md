# Scenario

**Feature**: wrk --web serves the React SPA (wrk-react)

```
# CLI surface
wrk --web [--port PORT] [--dev]
  -> bind 127.0.0.1:<port>
  -> stdout: http://127.0.0.1:<port>/\n
  -> serve GET / (SPA home) and client routes e.g. /mockup/repo-view
  -> mount wrkserver at /api/wrk/* (GET /api/wrk/projects, …)
  -> block until SIGINT/SIGTERM

# Invalid combinations
wrk --web --list / --status / positionals  -> non-zero; mutual exclusion / unexpected args
wrk --port PORT (no --web)                 -> non-zero; --port only valid with --web

# Help
wrk -h  -> documents --web and --port
```

## Preconditions

- Reuses root harness: session-built `wrk` binary, isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- `--web` needs no git checkout; leaves set `RepoDir = WorkRoot` unless noted.
- Long-running success leaves set `Request.WebProbe` + `WebPath` so root `Run` starts the server, HTTP GETs the path, then kills the process (never hang).
- Free TCP ports come from `freeLocalPort` (root SETUP) so parallel leaves do not collide.

## Steps

1. Root Setup creates `WorkRoot` / `WrkHome`.
2. Descendants set `Args` (and `WebProbe`/`WebPath` for serve leaves).
3. Root `Run` either runs wrk to completion (error/help leaves) or `runWebProbe` (serve leaves).

## Context

- Bind address is localhost only (`127.0.0.1`).
- Successful start prints a single stdout line: `http://127.0.0.1:<port>/` with trailing `\n`.
- HTML markers for tests: `task`, `changes`/`worktree`, `Main`, `Remote`, title/heading contains `wrk`.
- Empty projects API always returns JSON array (never null): `{"projects":[]}`.

```go
import "strconv"

func Setup(t *testing.T, req *Request) error {
	// --web does not require a git checkout; isolate under WorkRoot.
	req.RepoDir = req.WorkRoot
	return nil
}

// setupWebProbe configures a free --port and enables root Run's long-running
// HTTP probe for the given path (e.g. "/" or "/api/wrk/projects").
func setupWebProbe(t *testing.T, req *Request, webPath string) {
	t.Helper()
	port := freeLocalPort(t)
	req.Args = []string{"--web", "--port", strconv.Itoa(port)}
	req.WebProbe = true
	req.WebPath = webPath
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
}
```
