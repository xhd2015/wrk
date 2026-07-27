# Scenario

**Feature**: one module root → same items as PlanLocalReinstalls (M1)

```
# M1: single-element moduleRoots ≡ PlanLocalReinstalls(moduleRoot, binDir)
# fixture: ./cmd/present main + $binDir/present
PlanLocalReinstallsMulti([mod], binDir)
  -> Modules=[{ModuleName: multi-single, Items:[{present, go-install, ./cmd/present, install}]}]
# and Items match PlanLocalReinstalls for the same root
```

## Steps

1. Create module at `{WorkRoot}/mod` with path `example.com/multi-single`.
2. Write `./cmd/present` as `package main`.
3. Touch `$binDir/present`.
4. Set `ModuleRoots` to that single absolute path.
5. Expect one module block with one install item; Assert also compares to
   single-module `PlanLocalReinstalls` item set.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	modRoot := filepath.Join(req.WorkRoot, "mod")
	writeGoMod(t, modRoot, "example.com/multi-single")
	writePackageMain(t, filepath.Join(modRoot, "cmd", "present"))
	touchBin(t, req.BinDir, "present")

	req.ModuleRoots = []string{modRoot}
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: modRoot,
			ModuleName: "multi-single",
			Items: []WantPlanItem{
				{
					BinName: "present",
					Method:  methodGoInstall,
					RelPath: "./cmd/present",
					Action:  actionInstall,
				},
			},
		},
	}
	return nil
}
```
