# Scenario

**Feature**: ambiguous cmd and ambiguous script same bin → omit + both warnings

```
# S4g: two cmd + two script for bin foo; bin present
PlanLocalReinstalls
  -> Items=[]
  -> Diagnostics=[
       {warning, ambiguous-cmd, foo, [./cmd/foo, ./cmd/nested/foo]},
       {warning, ambiguous-script, foo, [./script/foo/install, ./script/x/foo/install]},
     ]
  # ordered by BinName then Kind (ambiguous-cmd < ambiguous-script)
```

## Steps

1. Write `go.mod` with module `example.com/both-amb`.
2. Write two cmd mains and two script installs for bin `foo`.
3. Touch `$binDir/foo`.
4. Expect empty Items and both warnings in Kind order.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/both-amb")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "x", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "both-amb"
	req.WantItems = []WantPlanItem{}
	req.WantDiagnostics = []WantDiagnostic{
		{
			Level:   diagLevelWarning,
			Kind:    diagKindAmbCmd,
			BinName: "foo",
			Paths:   []string{"./cmd/foo", "./cmd/nested/foo"},
		},
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
