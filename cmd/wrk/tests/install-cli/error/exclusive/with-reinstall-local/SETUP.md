# Scenario

**Feature**: --install is mutually exclusive with --reinstall-local

```
# I15: ./cmd/tool exists
mod/ -> wrk --install tool --reinstall-local
  -> non-zero; stderr names both flags
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-excl-reinstall`.
2. Write `./cmd/tool` as `package main`.
3. Run `wrk --install tool --reinstall-local`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-excl-reinstall")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--install", "tool", "--reinstall-local"}
	return nil
}
```
