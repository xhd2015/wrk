# Scenario

**Feature**: a named bin absent from GOBIN is planned for install, never skipped

```
# I1: ./cmd/tool exists; GOBIN has no tool
mod/ -> wrk --install tool --dry-run
  -> would: go install ./cmd/tool
  -> would: install 1 binaries
  -> no skip: line (the binDir gate is not consulted)
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-absent`.
2. Write `./cmd/tool` as `package main`.
3. Leave GOBIN empty (no stub).
4. Run `wrk --install tool --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-absent")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--install", "tool", "--dry-run"}
	return nil
}
```
