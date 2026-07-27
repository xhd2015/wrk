# Scenario

**Feature**: monorepo root requiring nested sibling is intra, not cross

```
# mono owns example.com/mono (.) and example.com/mono/sub (sub)
# root go.mod: require example.com/mono/sub v0.0.0
BuildInventory
  -> IntraEdges contains that require (consumer==owner project)
  -> CrossEdges does not
```

## Steps

1. Create monorepo with root module `example.com/mono` and nested `sub/` module
   `example.com/mono/sub`.
2. Root go.mod requires `example.com/mono/sub@v0.0.0` (intra sibling require).
3. Register only the monorepo path.
4. Expect IntraEdges has the edge; CrossEdges empty.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	monoPath := filepath.Join(req.WorkRoot, "repos", "mono")
	// initRootAndSub then rewrite root go.mod to add require on sub.
	initRootAndSubModuleRepo(t, monoPath, "example.com/mono")
	// Rewrite root go.mod with the intra require and recommit so scan sees it.
	writeGoMod(t, monoPath, "example.com/mono", "example.com/mono/sub@v0.0.0")
	runGitIsolated(t, monoPath, "add", "go.mod")
	runGitIsolated(t, monoPath, "commit", "-m", "root requires nested sub")

	monoPath = resolvePath(t, monoPath)
	writeProjectsJSON(t, req.WrkHome, monoPath)

	req.WantProjectPaths = []string{monoPath}
	req.WantModules = []WantModule{
		{ProjectPath: monoPath, Dir: ".", Path: "example.com/mono"},
		{ProjectPath: monoPath, Dir: "sub", Path: "example.com/mono/sub"},
	}
	req.WantCrossEdges = []WantEdge{}
	req.WantIntraEdges = []WantEdge{
		{
			ConsumerProject: monoPath,
			ConsumerModule:  "example.com/mono",
			DepPath:         "example.com/mono/sub",
			DepVersion:      "v0.0.0",
			OwnerProject:    monoPath,
		},
	}
	req.WantSkippedPaths = []string{}
	return nil
}
```
