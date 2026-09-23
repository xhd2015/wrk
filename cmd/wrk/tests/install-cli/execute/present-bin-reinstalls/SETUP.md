# Scenario

**Feature**: an existing stub is replaced by a real install (gate off, L2)

```
# I6: GOBIN/tool stub present; ./cmd/tool prints tool-ok
mod/ -> wrk --install tool      (in-process capture; real go install)
  -> installed 1, failed 0; stub replaced by a runnable binary
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-replace`.
2. Write `./cmd/tool` printing `tool-ok`.
3. Touch `$GOBIN/tool` (stub contents `stub-binary`).
4. Run `wrk --install tool` and assert the stub was replaced.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-replace")
	writePackageMainPrints(t, filepath.Join(req.ModuleRoot, "cmd", "tool"), "tool-ok")
	touchBin(t, req.BinDir, "tool")
	req.Args = []string{"--install", "tool"}
	return nil
}
```
