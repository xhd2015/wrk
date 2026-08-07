# Scenario

**Bug**: execute re-roots go install Dir+RelPath for package under nested cmd/go.mod (E1)

```
# agent-pro shape: root go.mod + nested cmd/go.mod; package main at cmd/foo
mod/
  go.mod example.com/cli-nested-cmd-root
  cmd/go.mod example.com/cli-nested-cmd-mod + foo/ prints "foo-ok"
  GOBIN/foo stub
  -> wrk --reinstall-local
  -> go install ./foo          # Dir=mod/cmd (nearest go.mod), not ./cmd/foo @ mod
  -> reinstalled 1, skipped 0, failed 0
  -> GOBIN/foo runs and prints foo-ok
```

## Steps

1. Write root module `example.com/cli-nested-cmd-root` with **no** root-owned
   package main outside the nested module.
2. Write nested module at `cmd/` as `example.com/cli-nested-cmd-mod` with
   package main at `foo/` that prints `foo-ok`.
3. Touch `$GOBIN/foo` stub so Action=install.
4. Run `wrk --reinstall-local` (no `--dry-run`) from ModuleRoot.
5. Expect execute re-root: progress `go install ./foo`, success summary, and a
   real binary that prints `foo-ok` (proves `go install` used the nested module
   as `Dir`, not the parent with unre-rooted `./cmd/foo`).

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
	writePackageMainPrints(t, filepath.Join(cmdMod, "foo"), "foo-ok")

	touchBin(t, req.BinDir, "foo")

	req.Args = []string{"--reinstall-local"}
	return nil
}
```
