# Scenario

**Feature**: same bin install×install from two modules → hard error (M3)

```
# M3: mod-a and mod-b both ./cmd/samebin + $binDir/samebin present
# → both Action=install → collision error naming bin + both modules
PlanLocalReinstallsMulti([mod-a, mod-b], binDir)
  -> error (mentions "samebin" and both module roots or names)
```

## Steps

1. Create `{WorkRoot}/mod-a` module `example.com/mod-a` with `./cmd/samebin`.
2. Create `{WorkRoot}/mod-b` module `example.com/mod-b` with `./cmd/samebin`.
3. Touch `$binDir/samebin` so both modules would Action=install.
4. Pass both absolute roots in ModuleRoots.
5. Expect non-nil error; substrings include bin name `samebin` and both
   module roots (or at least path bases `mod-a` and `mod-b`).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	modA := filepath.Join(req.WorkRoot, "mod-a")
	modB := filepath.Join(req.WorkRoot, "mod-b")

	writeGoMod(t, modA, "example.com/mod-a")
	writePackageMain(t, filepath.Join(modA, "cmd", "samebin"))

	writeGoMod(t, modB, "example.com/mod-b")
	writePackageMain(t, filepath.Join(modB, "cmd", "samebin"))

	touchBin(t, req.BinDir, "samebin")

	req.ModuleRoots = []string{modA, modB}
	req.WantError = true
	// Error must name the bin and identify both claiming modules.
	// Accept either absolute roots or path bases in the message.
	req.WantErrSubstrs = []string{
		"samebin",
		"mod-a",
		"mod-b",
	}
	req.WantModules = []WantModulePlan{}
	return nil
}
```
