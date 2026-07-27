# Scenario

**Feature**: other registered projects do not require source modules → empty would-update plan

```
# lib tagged v1.2.3; tool project requires only example.com/external@v9.9.9
cwd=lib -> wrk --propagate-tags --dry-run
  -> source: block
  -> no would: update module blocks
  -> footer 0 modules / 0 projects
```

## Steps

1. Create tagged lib `example.com/lib` at `v1.2.3`.
2. Create other project `tool` that does **not** require any source module.
3. Register both; dry-run from lib.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	toolPath := filepath.Join(req.WorkRoot, "repos", "tool")

	initSingleModuleRepo(t, libPath, "example.com/lib", nil)
	tagRepo(t, libPath, "v1.2.3")
	libPath = resolvePath(t, libPath)

	initSingleModuleRepo(t, toolPath, "example.com/tool", []string{
		"example.com/external@v9.9.9",
	})
	toolPath = resolvePath(t, toolPath)

	req.SourcePath = libPath
	req.OtherPath = toolPath
	writeProjectsJSON(t, req.WrkHome, libPath, toolPath)

	req.RepoDir = libPath
	req.Args = []string{"--propagate-tags", "--dry-run"}
	captureRepoSnapshots(t, req)
	return nil
}
```
