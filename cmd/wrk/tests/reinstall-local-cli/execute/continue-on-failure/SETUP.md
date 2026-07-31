# Scenario

**Feature**: execute continues after a failing install and still installs later bins (soft exit)

```
# E3: GOBIN has stubs for broken + good (lexicographic: broken then good)
# ./cmd/broken does not compile; ./cmd/good prints good-ok
mod/ -> wrk --reinstall-local
  -> go install ./cmd/broken  # fails (streamed go errors)
  -> go install ./cmd/good    # still runs
  -> reinstalled 1, skipped 0, failed 1
  -> exit 0; stderr warning: (reinstall/fail); GOBIN/good runs good-ok
```

## Steps

1. Write `go.mod` with module `example.com/cli-exec-partial`.
2. Write `./cmd/broken` as non-compiling `package main`.
3. Write `./cmd/good` as `package main` that prints `good-ok`.
4. Touch `$GOBIN/broken` and `$GOBIN/good` stubs so both are install actions.
5. Run `wrk --reinstall-local` (no `--dry-run`) from module root.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-partial")
	writeBrokenPackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "broken"))
	writePackageMainPrints(t, filepath.Join(req.ModuleRoot, "cmd", "good"), "good-ok")
	touchBin(t, req.BinDir, "broken")
	touchBin(t, req.BinDir, "good")
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
