# Scenario

**Feature**: cmd package main not present in binDir → skip action

```
# S2: ./cmd/missing main, no $binDir/missing
PlanLocalReinstalls
  -> Items=[{BinName:missing, Method:go-install, RelPath:./cmd/missing, Action:skip}]
```

## Steps

1. Write `go.mod` with module `example.com/cmd-absent`.
2. Write `./cmd/missing` as `package main`.
3. Do **not** create `$binDir/missing`.
4. Expect one skip item (still listed).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	writeGoMod(t, req.ModuleRoot, "example.com/cmd-absent")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "missing"))
	// intentionally no touchBin for "missing"

	req.WantModuleName = "cmd-absent"
	req.WantItems = []WantPlanItem{
		{
			BinName: "missing",
			Method:  methodGoInstall,
			RelPath: "./cmd/missing",
			Action:  actionSkip,
		},
	}
	return nil
}
```
