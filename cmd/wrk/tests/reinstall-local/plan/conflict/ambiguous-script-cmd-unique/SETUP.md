# Scenario

**Feature**: unique cmd + ambiguous script → cmd item; script warning only

```
# S4f: ./cmd/foo + ./script/foo/install + ./script/x/foo/install; bin present
PlanLocalReinstalls
  -> Items=[{BinName:foo, Method:go-install, RelPath:./cmd/foo, Action:install}]
  -> Diagnostics=[{Level:warning, Kind:ambiguous-script, BinName:foo,
       Paths:[./script/foo/install, ./script/x/foo/install]}]
```

## Steps

1. Write `go.mod` with module `example.com/amb-script-cmd`.
2. Write unique `./cmd/foo` and two script installs for bin `foo`.
3. Touch `$binDir/foo`.
4. Expect cmd survivor + ambiguous-script warning only.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeGoMod(t, req.ModuleRoot, "example.com/amb-script-cmd")
	writePackageMain(t, filepath.Join(req.ModuleRoot, "cmd", "foo"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "foo", "install"))
	writePackageMain(t, filepath.Join(req.ModuleRoot, "script", "x", "foo", "install"))
	touchBin(t, req.BinDir, "foo")

	req.WantModuleName = "amb-script-cmd"
	req.WantItems = []WantPlanItem{
		{
			BinName: "foo",
			Method:  methodGoInstall,
			RelPath: "./cmd/foo",
			Action:  actionInstall,
		},
	}
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
