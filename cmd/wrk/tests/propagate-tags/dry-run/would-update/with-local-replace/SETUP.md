# Scenario

**Feature**: outdated require with cross-project local replace plans drop-replace without writing

```
# lib tagged v1.2.3
# app requires example.com/lib@v1.0.0 + replace example.com/lib => <abs-lib>
cwd=lib -> wrk --propagate-tags --dry-run
  -> would: update example.com/app  (lib v1.0.0 -> v1.2.3)
  -> would: drop replace example.com/lib  (project app)
  -> app go.mod still has replace and old require after dry-run
```

## Steps

1. Create tagged `repos/lib` root module `example.com/lib` at `v1.2.3`.
2. Create `repos/app` with older require **and** local replace pointing at lib.
3. Register both; run dry-run from lib; snapshot go.mod/tags/HEAD first.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	initSingleModuleRepo(t, libPath, "example.com/lib", nil)
	tagRepo(t, libPath, "v1.2.3")
	libPath = resolvePath(t, libPath)

	// Local replace to absolute lib path (cross-project).
	initSingleModuleRepo(t, appPath, "example.com/app",
		[]string{"example.com/lib@v1.0.0"},
		"example.com/lib=>"+libPath,
	)
	appPath = resolvePath(t, appPath)

	req.SourcePath = libPath
	req.AppPath = appPath
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)

	req.RepoDir = libPath
	req.Args = []string{"--propagate-tags", "--dry-run"}
	captureRepoSnapshots(t, req)
	return nil
}
```
