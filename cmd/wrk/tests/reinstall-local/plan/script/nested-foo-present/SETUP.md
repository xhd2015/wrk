# Scenario

**Feature**: nested script install with bin present → go-run-install

```
# S3: ./script/foo/install + $binDir/foo
PlanLocalReinstalls
  -> Items=[{BinName:foo, Method:go-run-install, RelPath:./script/foo/install, Action:install}]
```

## Steps

1. Write `go.mod` with module `example.com/script-foo`.
2. Write `./script/foo/install` as `package main`.
3. Touch `$binDir/foo`.
4. Expect one go-run-install install item named `foo`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeGoMod(t, req.ModuleRoot, "example.com/script-foo")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "script-foo"
	req.WantItems = []WantPlanItem{
		{
			BinName: "foo",
			Method:  methodGoRunInstall,
			RelPath: "./script/foo/install",
			Action:  actionInstall,
		},
	}
	return nil
}
```
