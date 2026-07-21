# Scenario

**Feature**: multi-module execute continues after a failing install in an earlier module (E2)

```
# E2: root module broken bin fails; nested tools module still installs
mod/
  go.mod example.com/cli-exec-multi-partial-root + cmd/broken (does not compile)
  tools/go.mod example.com/cli-exec-multi-partial-tools + cmd/toolgood (prints toolgood-ok)
  GOBIN/{broken,toolgood} stubs
  -> wrk --reinstall-local
  -> go install ./cmd/broken     # fails (streamed go errors)
  -> go install ./cmd/toolgood   # still runs (later module)
  -> reinstalled 1, skipped 0, failed 1
  -> exit 1; GOBIN/toolgood runs toolgood-ok
```

## Steps

1. Write root module `example.com/cli-exec-multi-partial-root` with
   `./cmd/broken` as non-compiling `package main`.
2. Write nested `tools/` module `example.com/cli-exec-multi-partial-tools` with
   `./cmd/toolgood` as `package main` that prints `toolgood-ok`.
3. Touch `$GOBIN/broken` and `$GOBIN/toolgood` stubs so both are install actions.
4. Run `wrk --reinstall-local` (no `--dry-run`) from ModuleRoot.
5. Expect continue-on-failure across module boundaries; later module still installs.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cli-exec-multi-partial-root")
	writeBrokenPackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "broken"))

	toolsMod := filepath.Join(req.ModuleRoot, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-exec-multi-partial-tools")
	writePackageMainPrints(t, filepath.Join(toolsMod, "cmd", "toolgood"), "toolgood-ok")

	touchBin(t, req.BinDir, "broken")
	touchBin(t, req.BinDir, "toolgood")

	req.Args = []string{"--reinstall-local"}
	return nil
}
```
