# Scenario

**Feature**: non-git dir; walk-up from subdir finds go.mod → single module (R2)

```
# R2: plain directory tree (no .git); workDir under the module
mod/
  go.mod + cmd/onlybin
  nested/   <- workDir
  -> ResolveReinstallScanRoot -> mod
  -> PlanLocalReinstallsFromWorkDir
  -> Modules: [mod onlybin install]
```

## Steps

1. Create module `{WorkRoot}/mod` path `example.com/nongit-mod` with
   `./cmd/onlybin` package main (no git init).
2. Create subdirectory `{mod}/nested` as WorkDir.
3. Touch `$binDir/onlybin`.
4. Expect ScanRoot=mod; one module plan with install onlybin.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	modRoot := filepath.Join(req.WorkRoot, "mod")
	writeGoMod(t, modRoot, "example.com/nongit-mod")
	writePackageMain(t, filepath.Join(modRoot, "cmd", "onlybin"))

	nested := filepath.Join(modRoot, "nested")
	mkdirAll(t, nested)

	modRoot = resolvePath(t, modRoot)
	nested = resolvePath(t, nested)

	touchBin(t, req.BinDir, "onlybin")

	req.WorkDir = nested
	req.WantScanRoot = modRoot
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: modRoot,
			ModuleName: "nongit-mod",
			Items: []WantPlanItem{
				{
					BinName: "onlybin",
					Method:  methodGoInstall,
					RelPath: "./cmd/onlybin",
					Action:  actionInstall,
				},
			},
		},
	}
	return nil
}
```
