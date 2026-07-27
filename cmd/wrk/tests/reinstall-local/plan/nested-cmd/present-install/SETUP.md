# Scenario

**Feature**: nested ./cmd/nested/tool present in bin → go-install tool

```
# S6: ./cmd/nested/tool + $binDir/tool
PlanLocalReinstalls
  -> Items=[{BinName:tool, Method:go-install, RelPath:./cmd/nested/tool, Action:install}]
```

## Steps

1. Write `go.mod` with module `example.com/nested-cmd`.
2. Write `./cmd/nested/tool` as `package main`.
3. Touch `$binDir/tool` (not `nested`).
4. Expect one install item for bin `tool`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeGoMod(t, req.ModuleRoot, "example.com/nested-cmd")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "nested", "tool"))
	touchBin(t, req.BinDir, "tool")

	req.WantModuleName = "nested-cmd"
	req.WantItems = []WantPlanItem{
		{
			BinName: "tool",
			Method:  methodGoInstall,
			RelPath: "./cmd/nested/tool",
			Action:  actionInstall,
		},
	}
	return nil
}
```
