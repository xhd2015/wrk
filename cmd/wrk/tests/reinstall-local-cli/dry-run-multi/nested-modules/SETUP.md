# Scenario

**Feature**: multi-module dry-run prints headers + both would lines + across 2 (C1)

```
# C1: ModuleRoot with root go.mod + nested tools/go.mod; both bins present
mod/
  go.mod example.com/cli-multi-root + cmd/rootbin
  tools/go.mod example.com/cli-multi-tools + cmd/toolbin
  GOBIN/{rootbin,toolbin}
  -> wrk --reinstall-local --dry-run
  -> # module headers for both; would: go install each; across 2 modules
```

## Steps

1. Write root module `example.com/cli-multi-root` with `./cmd/rootbin` package main.
2. Write nested module `tools/` as `example.com/cli-multi-tools` with
   `./cmd/toolbin` package main.
3. Touch `$GOBIN/rootbin` and `$GOBIN/toolbin`.
4. Run from ModuleRoot (non-git walk-up scan root = ModuleRoot; scan finds both).
5. Expect multi dry-run format with K=2.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-multi-root")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "rootbin"))

	toolsMod := filepath.Join(req.ModuleRoot, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-multi-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
