# Scenario

**Feature**: dry-run with present bin prints would: go install

```
# C1: ./cmd/present + GOBIN/present
mod/ -> wrk --reinstall-local --dry-run
  -> would: go install ./cmd/present
  -> would: reinstall 1 binaries (0 skipped)
```

## Steps

1. Write `go.mod` with module `example.com/cli-present`.
2. Write `./cmd/present` as `package main`.
3. Touch `$GOBIN/present` stub.
4. Run `wrk --reinstall-local --dry-run` from module root.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-present")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "present"))
	touchBin(t, req.BinDir, "present")
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
