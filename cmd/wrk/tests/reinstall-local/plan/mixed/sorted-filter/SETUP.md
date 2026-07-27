# Scenario

**Feature**: mixed present/absent candidates yield sorted install and skip items

```
# S7: alpha (cmd, present), mid (script, absent), zed (cmd, present)
# discovery order intentionally not sorted; plan must sort by BinName
PlanLocalReinstalls
  -> Items=[
       {alpha, go-install, ./cmd/alpha, install},
       {mid, go-run-install, ./script/mid/install, skip},
       {zed, go-install, ./cmd/zed, install},
     ]
```

## Steps

1. Write `go.mod` with module `example.com/mixed-plan`.
2. Create `./cmd/zed`, `./cmd/alpha` package mains (write zed first to scramble
   discovery order vs sort order).
3. Create `./script/mid/install` package main.
4. Touch only `$binDir/alpha` and `$binDir/zed` (not `mid`).
5. Expect three items sorted alpha < mid < zed with correct actions.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeGoMod(t, req.ModuleRoot, "example.com/mixed-plan")
	// Write in non-sorted order to ensure product sorts by BinName.
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "zed"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "mid", "install"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "alpha"))
	touchBin(t, req.BinDir, "alpha")
	touchBin(t, req.BinDir, "zed")
	// mid intentionally absent from binDir

	req.WantModuleName = "mixed-plan"
	req.WantItems = []WantPlanItem{
		{
			BinName: "alpha",
			Method:  methodGoInstall,
			RelPath: "./cmd/alpha",
			Action:  actionInstall,
		},
		{
			BinName: "mid",
			Method:  methodGoRunInstall,
			RelPath: "./script/mid/install",
			Action:  actionSkip,
		},
		{
			BinName: "zed",
			Method:  methodGoInstall,
			RelPath: "./cmd/zed",
			Action:  actionInstall,
		},
	}
	return nil
}
```
