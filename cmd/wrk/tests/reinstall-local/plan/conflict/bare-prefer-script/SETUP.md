# Scenario

**Feature**: unique cmd + bare script install same ModuleName bin → prefer-script notice

```
# S4b: module example.com/demo; ./cmd/demo + ./script/install; $binDir/demo
PlanLocalReinstalls
  -> Items=[{BinName:demo, Method:go-run-install, RelPath:./script/install, Action:install}]
  -> Diagnostics=[{Level:notice, Kind:prefer-script, BinName:demo,
       Paths:[./cmd/demo, ./script/install]}]
```

## Steps

1. Write `go.mod` with module `example.com/demo` (basename `demo`).
2. Write `./cmd/demo` as `package main`.
3. Write bare `./script/install` as `package main`.
4. Touch `$binDir/demo`.
5. Expect script item + prefer-script notice (paths sorted).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/demo")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "demo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "install"))
	touchBin(t, req.BinDir, "demo")

	req.WantModuleName = "demo"
	req.WantItems = []WantPlanItem{
		{
			BinName: "demo",
			Method:  methodGoRunInstall,
			RelPath: "./script/install",
			Action:  actionInstall,
		},
	}
	req.WantDiagnostics = []WantDiagnostic{
		{
			Level:   diagLevelNotice,
			Kind:    diagKindPrefer,
			BinName: "demo",
			Paths:   []string{"./cmd/demo", "./script/install"},
		},
	}
	return nil
}
```
