# Scenario

**Feature**: cmd package main present in binDir → install action

```
# S1: ./cmd/present + $binDir/present exists
PlanLocalReinstalls
  -> Items=[{BinName:present, Method:go-install, RelPath:./cmd/present, Action:install}]
```

## Steps

1. Write `go.mod` with module `example.com/cmd-present`.
2. Write `./cmd/present` as `package main`.
3. Touch `$binDir/present`.
4. Expect one install item.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeGoMod(t, req.ModuleRoot, "example.com/cmd-present")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "present"))
	touchBin(t, req.BinDir, "present")

	req.WantModuleName = "cmd-present"
	req.WantItems = []WantPlanItem{
		{
			BinName: "present",
			Method:  methodGoInstall,
			RelPath: "./cmd/present",
			Action:  actionInstall,
		},
	}
	return nil
}
```
