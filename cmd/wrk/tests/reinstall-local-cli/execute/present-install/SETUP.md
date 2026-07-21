# Scenario

**Feature**: execute reinstalls a present cmd binary via go install

```
# E1: ./cmd/tool package main (prints tool-v1) + GOBIN/tool stub
mod/ -> wrk --reinstall-local
  -> go install ./cmd/tool
  -> reinstalled 1, skipped 0, failed 0
  -> GOBIN/tool is a real executable that prints tool-v1
```

## Steps

1. Write `go.mod` with module `example.com/cli-exec-tool`.
2. Write `./cmd/tool` as `package main` that prints `tool-v1`.
3. Touch `$GOBIN/tool` stub so Action=install (filter requires present bin).
4. Run `wrk --reinstall-local` (no `--dry-run`) from module root.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-tool")
	writePackageMainPrints(t, filepath.Join(req.ModuleRoot, "cmd", "tool"), "tool-v1")
	touchBin(t, req.BinDir, "tool")
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
