# Scenario

**Feature**: a successful install dry-run records events.jsonl command "install"

```
# I16: ./cmd/tool exists; GOBIN empty
mod/ -> wrk --install tool --dry-run
  -> exit 0; last event: command=install, exit_code=0
  -> args include --install, tool, --dry-run
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-events`.
2. Write `./cmd/tool` as `package main` (forced plan: no stub needed).
3. Run `wrk --install tool --dry-run`.
4. Assert the last `events.jsonl` event (do not re-invoke wrk before reading).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-events")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--install", "tool", "--dry-run"}
	return nil
}
```
