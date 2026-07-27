# Scenario

**Feature**: dry-run with only missing bins prints skip lines and N=0 summary

```
# C2: ./cmd/missing, no GOBIN/missing
mod/ -> wrk --reinstall-local --dry-run
  -> skip: missing (not in <gobin>)
  -> would: reinstall 0 binaries (1 skipped)
```

## Steps

1. Write `go.mod` with module `example.com/cli-skip`.
2. Write `./cmd/missing` as `package main`.
3. Do **not** create `$GOBIN/missing`.
4. Run `wrk --reinstall-local --dry-run` from module root.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-skip")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "missing"))
	// intentionally no touchBin for "missing"
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
