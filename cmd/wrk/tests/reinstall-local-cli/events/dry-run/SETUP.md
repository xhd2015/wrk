# Scenario

**Feature**: dry-run success records events.jsonl command "reinstall-local"

```
# H1: wrk --reinstall-local --dry-run success -> events.jsonl last event
mod/ -> wrk --reinstall-local --dry-run
  -> exit 0
  -> last event: command=reinstall-local, exit_code=0
  -> args include --reinstall-local and --dry-run
```

## Steps

1. Write `go.mod` with module `example.com/cli-events-dry`.
2. Write `./cmd/missing` as `package main` (skip-only plan; no install work).
3. Run `wrk --reinstall-local --dry-run` from module root.
4. Assert last `events.jsonl` event (do not re-invoke wrk before read).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-events-dry")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "missing"))
	// skip-only plan: no GOBIN stub — keeps dry-run fast and exit 0
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
