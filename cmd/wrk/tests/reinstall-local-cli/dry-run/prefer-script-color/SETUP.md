# Scenario

**Feature**: --color colors notice: prefix grey on prefer-script stderr line

```
# C3-nc: same fixtures as script-wins + --color
mod/ -> wrk --reinstall-local --dry-run --color
  -> stdout plan plain (no ANSI)
  -> stderr: <grey>notice:</grey> bin foo: preferring ./script/foo/install over ./cmd/foo
```

## Steps

1. Write cmd+script conflict fixtures for bin `foo`.
2. Touch `$GOBIN/foo`.
3. Run with `--reinstall-local --dry-run --color`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-prefer-color")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	touchBin(t, req.BinDir, "foo")
	req.Args = []string{"--reinstall-local", "--dry-run", "--color"}
	return nil
}
```
