# Scenario

**Feature**: real product binary installs the named bin into GOBIN (L3 e2e)

```
# I5: ./cmd/tool prints tool-ok; GOBIN empty
mod/ -> wrk --install tool      (process boundary: product binary)
  -> go install ./cmd/tool
  -> installed 1, failed 0
  -> GOBIN/tool exists, is executable, prints tool-ok
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-exec`.
2. Write `./cmd/tool` printing `tool-ok`.
3. Run the session `wrk` binary with `--install tool` (no `--dry-run`).
4. Assert the installed binary runs and the summary counts it once.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// L3: real product binary + real go install (label: e2e).
	req.InProcess = false
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-exec")
	writePackageMainPrints(t, filepath.Join(req.ModuleRoot, "cmd", "tool"), "tool-ok")
	req.Args = []string{"--install", "tool"}
	return nil
}
```
