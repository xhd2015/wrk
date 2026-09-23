# Scenario

**Feature**: an already-installed bin is still a forced install (no skip)

```
# I2: ./cmd/tool exists; GOBIN/tool stub exists
mod/ -> wrk --install tool --dry-run
  -> would: go install ./cmd/tool (forced; presence in GOBIN is irrelevant)
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-present`.
2. Write `./cmd/tool` as `package main`.
3. Touch `$GOBIN/tool` (stub).
4. Run `wrk --install tool --dry-run`; expect the same plan as an absent bin.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-present")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "tool"))
	touchBin(t, req.BinDir, "tool")
	req.Args = []string{"--install", "tool", "--dry-run"}
	return nil
}
```
