# Scenario

**Feature**: two script installs share bin name → omit bin + ambiguous-script warning

```
# S4d: ./script/foo/install + ./script/x/foo/install, no cmd, bin present
PlanLocalReinstalls
  -> Items=[]
  -> Diagnostics=[{Level:warning, Kind:ambiguous-script, BinName:foo,
       Paths:[./script/foo/install, ./script/x/foo/install]}]
```

## Steps

1. Write `go.mod` with module `example.com/amb-script`.
2. Write `./script/foo/install` and `./script/x/foo/install` as `package main`.
3. Touch `$binDir/foo`.
4. Expect empty Items and one ambiguous-script warning with sorted paths.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/amb-script")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "x", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "amb-script"
	req.WantItems = []WantPlanItem{}
	req.WantDiagnostics = []WantDiagnostic{
		{
			Level:   diagLevelWarning,
			Kind:    diagKindAmbScript,
			BinName: "foo",
			Paths:   []string{"./script/foo/install", "./script/x/foo/install"},
		},
	}
	return nil
}
```
