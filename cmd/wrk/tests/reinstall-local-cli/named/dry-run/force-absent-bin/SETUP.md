# Scenario

**Feature**: named mode installs even when the bin is not in GOBIN

```
# ./cmd/tool exists; GOBIN has no tool stub
mod/ -> wrk --reinstall-local tool --dry-run
  -> would: go install ./cmd/tool (not skip:)
```

## Steps

1. Write module with `./cmd/tool` package main.
2. Leave GOBIN empty (no stub).
3. Run `--reinstall-local tool --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/named-force")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	req.Args = []string{"--reinstall-local", "tool", "--dry-run"}
	return nil
}
```
