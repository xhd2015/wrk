# Scenario

**Feature**: multi-module dry-run install×install collision is a hard error (C3)

```
# C3: parent scan root with two nested modules both ./cmd/samebin + GOBIN/samebin
mod/
  go.mod (parent scan root; no install candidates required)
  mod-a/go.mod + cmd/samebin
  mod-b/go.mod + cmd/samebin
  GOBIN/samebin present
  -> wrk --reinstall-local --dry-run
  -> non-zero; stderr names bin samebin (and identifies both modules)
```

## Steps

1. Write parent module `example.com/cli-coll-parent` at ModuleRoot (scan root).
2. Write nested `mod-a` module `example.com/cli-coll-a` with `./cmd/samebin`.
3. Write nested `mod-b` module `example.com/cli-coll-b` with `./cmd/samebin`.
4. Touch `$GOBIN/samebin` so both nested modules would Action=install.
5. Run from ModuleRoot; expect non-zero exit and stderr naming the bin.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-coll-parent")

	modA := filepath.Join(req.ModuleRoot, "mod-a")
	writeGoMod(t, modA, "example.com/cli-coll-a")
	writePackageMain(t, filepath.Join(modA, "cmd", "samebin"))

	modB := filepath.Join(req.ModuleRoot, "mod-b")
	writeGoMod(t, modB, "example.com/cli-coll-b")
	writePackageMain(t, filepath.Join(modB, "cmd", "samebin"))

	touchBin(t, req.BinDir, "samebin")

	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
