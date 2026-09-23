# Scenario

**Feature**: only the requested names are planned (names are arbitrary args)

```
# I3: ./cmd/tool + ./cmd/other; request only tool
mod/ -> wrk --install tool --dry-run
  -> would: go install ./cmd/tool only (other is not planned)
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-select`.
2. Write `./cmd/tool` and `./cmd/other` as `package main`.
3. Run `wrk --install tool --dry-run` (the name is an arbitrary arg after the flag).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-select")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "other"))
	req.Args = []string{"--install", "tool", "--dry-run"}
	return nil
}
```
