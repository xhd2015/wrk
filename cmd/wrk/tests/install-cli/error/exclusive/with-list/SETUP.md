# Scenario

**Feature**: --install is mutually exclusive with --list

```
# I14: ./cmd/tool exists
mod/ -> wrk --install tool --list
  -> non-zero; stdout empty; stderr mentions mutual exclusion
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-excl-list`.
2. Write `./cmd/tool` as `package main`.
3. Run `wrk --install tool --list`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-excl-list")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--install", "tool", "--list"}
	return nil
}
```
