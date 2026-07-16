# Scenario

**Feature**: two cmd package mains share bin name → omit bin + ambiguous-cmd warning

```
# S4c: ./cmd/foo + ./cmd/nested/foo, bin present, no script
PlanLocalReinstalls
  -> Items=[]  (ambiguous-only bins omitted; not a skip row)
  -> Diagnostics=[{Level:warning, Kind:ambiguous-cmd, BinName:foo,
       Paths:[./cmd/foo, ./cmd/nested/foo]}]
```

## Steps

1. Write `go.mod` with module `example.com/amb-cmd`.
2. Write `./cmd/foo` and `./cmd/nested/foo` as `package main` (same BinName `foo`).
3. Touch `$binDir/foo` (would be install if unique — still omitted).
4. Expect empty Items and one ambiguous-cmd warning with sorted paths.

```go
func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/amb-cmd")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "foo"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "amb-cmd"
	req.WantItems = []WantPlanItem{}
	req.WantDiagnostics = []WantDiagnostic{
		{
			Level:   diagLevelWarning,
			Kind:    diagKindAmbCmd,
			BinName: "foo",
			Paths:   []string{"./cmd/foo", "./cmd/nested/foo"},
		},
	}
	return nil
}
```
