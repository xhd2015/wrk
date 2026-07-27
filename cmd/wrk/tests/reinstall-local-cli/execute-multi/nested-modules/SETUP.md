# Scenario

**Feature**: multi-module execute installs both root and nested bins into GOBIN (E1)

```
# E1: ModuleRoot with root go.mod + nested tools/go.mod; both bins present
mod/
  go.mod example.com/cli-exec-multi-root + cmd/rootbin (prints rootbin-ok)
  tools/go.mod example.com/cli-exec-multi-tools + cmd/toolbin (prints toolbin-ok)
  GOBIN/{rootbin,toolbin} stubs
  -> wrk --reinstall-local
  -> go install ./cmd/rootbin   # Dir=mod
  -> go install ./cmd/toolbin   # Dir=mod/tools
  -> reinstalled 2, skipped 0, failed 0
  -> both GOBIN bins run
```

## Steps

1. Write root module `example.com/cli-exec-multi-root` with `./cmd/rootbin`
   package main that prints `rootbin-ok`.
2. Write nested module `tools/` as `example.com/cli-exec-multi-tools` with
   `./cmd/toolbin` package main that prints `toolbin-ok`.
3. Touch `$GOBIN/rootbin` and `$GOBIN/toolbin` stubs so both are install actions.
4. Run `wrk --reinstall-local` (no `--dry-run`) from ModuleRoot.
5. Expect both installs and both bins executable under GOBIN.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-multi-root")
	writePackageMainPrints(t, filepath.Join(req.ModuleRoot, "cmd", "rootbin"), "rootbin-ok")

	toolsMod := filepath.Join(req.ModuleRoot, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-exec-multi-tools")
	writePackageMainPrints(t, filepath.Join(toolsMod, "cmd", "toolbin"), "toolbin-ok")

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	req.Args = []string{"--reinstall-local"}
	return nil
}
```
