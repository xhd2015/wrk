# Scenario

**Feature**: missing registry path is soft-skipped; good project still inventoried

```
# projects.json = [does-not-exist, good-lib]
BuildInventory
  -> SkippedPaths includes missing path
  -> Projects/Modules include only good-lib
  -> err == nil (soft-skip, not hard failure)
```

## Steps

1. Create one good project with module `example.com/good`.
2. Seed projects.json with a non-existent path **and** the good project path.
3. Expect SkippedPaths = [missing]; ProjectPaths = [good]; modules for good only.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	goodPath := filepath.Join(req.WorkRoot, "repos", "good")
	initSingleModuleRepo(t, goodPath, "example.com/good")
	goodPath = resolvePath(t, goodPath)

	missingPath := filepath.Join(req.WorkRoot, "does-not-exist", "gone")
	// Do not create missingPath on disk.
	writeProjectsJSON(t, req.WrkHome, missingPath, goodPath)

	req.WantProjectPaths = []string{goodPath}
	req.WantModules = []WantModule{
		{ProjectPath: goodPath, Dir: ".", Path: "example.com/good"},
	}
	req.WantCrossEdges = []WantEdge{}
	req.WantIntraEdges = []WantEdge{}
	req.WantSkippedPaths = []string{missingPath}
	return nil
}
```
