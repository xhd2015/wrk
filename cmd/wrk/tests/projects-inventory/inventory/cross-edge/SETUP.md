# Scenario

**Feature**: cross-project require becomes one CrossEdge when dep owner is registered

```
# app requires example.com/lib@v1.0.0; lib registered owns that module
BuildInventory
  -> CrossEdges = [{
       ConsumerProject=app, ConsumerModule=example.com/app,
       DepPath=example.com/lib, DepVersion=v1.0.0, OwnerProject=lib
     }]
  -> IntraEdges does not include that require
```

## Steps

1. Create lib repo with root module `example.com/lib`.
2. Create app repo with root module `example.com/app` requiring `example.com/lib@v1.0.0`.
3. Register both projects.
4. Expect exactly one cross edge; no intra edges for that require.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")
	initSingleModuleRepo(t, libPath, "example.com/lib")
	initSingleModuleRepo(t, appPath, "example.com/app", "example.com/lib@v1.0.0")

	libPath = resolvePath(t, libPath)
	appPath = resolvePath(t, appPath)
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)

	req.WantProjectPaths = []string{libPath, appPath}
	req.WantModules = []WantModule{
		{ProjectPath: libPath, Dir: ".", Path: "example.com/lib"},
		{ProjectPath: appPath, Dir: ".", Path: "example.com/app"},
	}
	req.WantCrossEdges = []WantEdge{
		{
			ConsumerProject: appPath,
			ConsumerModule:  "example.com/app",
			DepPath:         "example.com/lib",
			DepVersion:      "v1.0.0",
			OwnerProject:    libPath,
		},
	}
	req.WantIntraEdges = []WantEdge{}
	req.WantSkippedPaths = []string{}
	return nil
}
```
