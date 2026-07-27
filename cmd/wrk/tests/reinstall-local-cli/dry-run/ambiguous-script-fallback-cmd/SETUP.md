# Scenario

**Feature**: ambiguous script + unique cmd → dry-run falls back to go install cmd

```
# C3-fb: ./cmd/foo + two script installs for foo + GOBIN/foo
mod/ -> wrk --reinstall-local --dry-run
  -> stdout: would: go install ./cmd/foo
  -> stdout: would: reinstall 1 binaries (0 skipped)
  -> stderr: warning: bin foo: ambiguous under script (...); skipping
```

## Steps

1. Write `go.mod` with module `example.com/cli-amb-script-fb`.
2. Write unique `./cmd/foo` and two script installs sharing bin `foo`.
3. Touch `$GOBIN/foo`.
4. Run `wrk --reinstall-local --dry-run` from module root.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-amb-script-fb")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "x", "foo", "install"))
	touchBin(t, req.BinDir, "foo")
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
