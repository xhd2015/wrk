# Scenario

**Feature**: unknown name is a hard error

```
mod/ with ./cmd/only -> wrk --reinstall-local nope --dry-run
  -> non-zero; stderr mentions no install candidate / nope
```

## Steps

1. Write `./cmd/only`.
2. Run `--reinstall-local nope --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/named-miss")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "only"))
	req.Args = []string{"--reinstall-local", "nope", "--dry-run"}
	return nil
}
```
