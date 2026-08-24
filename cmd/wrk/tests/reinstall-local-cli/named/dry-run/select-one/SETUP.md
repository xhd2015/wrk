# Scenario

**Feature**: named mode keeps only the requested bin when several candidates exist

```
# ./cmd/keep + ./cmd/drop; GOBIN stubs both
mod/ -> wrk --reinstall-local keep --dry-run
  -> only would: go install ./cmd/keep
```

## Steps

1. Write `./cmd/keep` and `./cmd/drop`.
2. Touch both stubs under GOBIN.
3. Run `--reinstall-local keep --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/named-select")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "keep"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "drop"))
	touchBin(t, req.BinDir, "keep")
	touchBin(t, req.BinDir, "drop")
	req.Args = []string{"--reinstall-local", "keep", "--dry-run"}
	return nil
}
```
