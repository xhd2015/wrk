# Scenario

**Feature**: execute path with only missing bins skips without running go

```
# E2: ./cmd/missing package main, no GOBIN/missing
mod/ -> wrk --reinstall-local
  -> skip: missing (not in <gobin>)
  -> reinstalled 0, skipped 1, failed 0
  # must NOT run go install / go run
```

## Steps

1. Write `go.mod` with module `example.com/cli-exec-skip`.
2. Write `./cmd/missing` as `package main`.
3. Do **not** create `$GOBIN/missing`.
4. Run `wrk --reinstall-local` (no `--dry-run`) from module root.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-skip")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "missing"))
	// intentionally no touchBin for "missing"
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
