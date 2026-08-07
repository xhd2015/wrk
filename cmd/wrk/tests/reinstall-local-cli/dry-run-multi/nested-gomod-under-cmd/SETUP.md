# Scenario

**Bug**: dry-run re-roots go install path when package lives under nested cmd/go.mod (D1)

```
# agent-pro shape: root go.mod + nested cmd/go.mod; package main at cmd/foo
# discovery may still plan RelPath=./cmd/foo under the root module
mod/
  go.mod example.com/cli-nested-cmd-root
  cmd/go.mod example.com/cli-nested-cmd-mod + foo/ (package main)
  GOBIN/foo stub
  -> wrk --reinstall-local --dry-run
  -> would: go install ./foo   # post-re-root (not ./cmd/foo)
  # module headers may still name the plan (parent) module
```

## Steps

1. Write root module `example.com/cli-nested-cmd-root` with **no** root-owned
   package main outside the nested module.
2. Write nested module at `cmd/` as `example.com/cli-nested-cmd-mod` with
   `./foo` package main (path from scan root: `cmd/foo`).
3. Touch `$GOBIN/foo` so Action=install.
4. Run from ModuleRoot (scan finds root + nested `cmd/` modules).
5. Expect dry-run path strings to use **post-re-root** RelPath `./foo` (nearest
   go.mod under plan ModuleRoot is `cmd/`), not unre-rooted `./cmd/foo`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-nested-cmd-root")

	cmdMod := filepath.Join(req.ModuleRoot, "cmd")
	writeGoMod(t, cmdMod, "example.com/cli-nested-cmd-mod")
	writePackageMain(t, filepath.Join(cmdMod, "foo"))

	touchBin(t, req.BinDir, "foo")

	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
