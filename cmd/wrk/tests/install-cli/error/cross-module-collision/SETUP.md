# Scenario

**Feature**: a name claimed by two modules is a hard error

```
# I13: parent scan root with two nested modules both ./cmd/same
mod/
  go.mod (parent scan root)
  mod-a/go.mod + cmd/same
  mod-b/go.mod + cmd/same
  -> wrk --install same
  -> non-zero; stderr names bin same and both modules
```

## Steps

1. Write parent module `example.com/cli-install-coll-parent` at ModuleRoot.
2. Write nested `mod-a` and `mod-b`, each with `./cmd/same`.
3. Run `wrk --install same`; expect non-zero before any install.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeGoMod(t, req.ModuleRoot, "example.com/cli-install-coll-parent")

	modA := filepath.Join(req.ModuleRoot, "mod-a")
	writeGoMod(t, modA, "example.com/cli-install-coll-a")
	writePackageMain(t, filepath.Join(modA, "cmd", "same"))

	modB := filepath.Join(req.ModuleRoot, "mod-b")
	writeGoMod(t, modB, "example.com/cli-install-coll-b")
	writePackageMain(t, filepath.Join(modB, "cmd", "same"))

	req.Args = []string{"--install", "same"}
	return nil
}
```
