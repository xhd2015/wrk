# Scenario

**Feature**: multi-module scan lists root + nested modules across registered projects

```
# two registered projects
lib: example.com/lib (Dir=.) + example.com/lib/sub (Dir=sub)
app: example.com/app (Dir=.)
BuildInventory -> all three modules with correct Dir/Path/ownership
```

## Steps

1. Create `repos/lib` with root module `example.com/lib` and nested `sub/` → `example.com/lib/sub`.
2. Create `repos/app` with root module `example.com/app` (no nested modules).
3. Register both paths in projects.json.
4. Expect two projects and three modules with Dir/Path as above.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")
	initRootAndSubModuleRepo(t, libPath, "example.com/lib")
	initSingleModuleRepo(t, appPath, "example.com/app")

	libPath = resolvePath(t, libPath)
	appPath = resolvePath(t, appPath)
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)

	req.WantProjectPaths = []string{libPath, appPath}
	req.WantModules = []WantModule{
		{ProjectPath: libPath, Dir: ".", Path: "example.com/lib"},
		{ProjectPath: libPath, Dir: "sub", Path: "example.com/lib/sub"},
		{ProjectPath: appPath, Dir: ".", Path: "example.com/app"},
	}
	req.WantCrossEdges = []WantEdge{}
	req.WantIntraEdges = []WantEdge{}
	req.WantSkippedPaths = []string{}
	return nil
}
```
