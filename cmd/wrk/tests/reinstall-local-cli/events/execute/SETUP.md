# Scenario

**Feature**: execute success records events.jsonl command "reinstall-local"

```
# H-exec: wrk --reinstall-local (no --dry-run) success -> events.jsonl last event
mod/ -> wrk --reinstall-local
  -> exit 0
  -> last event: command=reinstall-local, exit_code=0
  -> args include --reinstall-local (not --dry-run)
```

## Steps

1. Write `go.mod` with module `example.com/cli-events-exec`.
2. Write `./cmd/missing` as `package main` (skip-only execute; no go install).
3. Run `wrk --reinstall-local` from module root.
4. Assert last `events.jsonl` event (do not re-invoke wrk before read).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-events-exec")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "missing"))
	// skip-only execute: no GOBIN stub — exit 0 without go install
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
