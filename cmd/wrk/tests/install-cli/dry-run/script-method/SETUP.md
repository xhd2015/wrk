# Scenario

**Feature**: a script/ install candidate resolves to `go run` (shared discovery)

```
# I4: only script/tool/install (package main); no cmd/tool
mod/ -> wrk --install tool --dry-run
  -> would: go run ./script/tool/install
```

## Steps

1. Write `go.mod` with module `example.com/cli-install-script`.
2. Write `./script/tool/install` as `package main`.
3. Run `wrk --install tool --dry-run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-script")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "tool", "install"))
	req.Args = []string{"--install", "tool", "--dry-run"}
	return nil
}
```
