# Scenario

**Feature**: multi-project nested modules with one cross-project require edge

```
# lib: example.com/lib (.) + example.com/lib/sub (sub)
# app: example.com/app (.) requires example.com/lib@v1.2.3
#      also requires example.com/external@v9.9.9 (unknown owner — not shown)
projects sorted by path: app then lib
wrk --projects-dep-graph
  -> both projects, three modules
  -> one cross-edge under app: → example.com/lib@v1.2.3  [lib]
  -> no edge for external
  -> footer 2 projects · 3 modules · 1 cross-edge
```

## Steps

1. Create `repos/lib` with root + nested `sub/` modules.
2. Create `repos/app` requiring `example.com/lib@v1.2.3` and an external module.
3. Register both paths; rely on path sort so `app` prints before `lib`.
4. Expect nested module lines and exactly one `→` line with owner basename `lib`.

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
	initRootAndSubModuleRepo(t, libPath, "example.com/lib")
	// Cross edge to registered lib; external require must not appear as an edge.
	initSingleModuleRepo(t, appPath, "example.com/app",
		"example.com/lib@v1.2.3",
		"example.com/external@v9.9.9",
	)
	libPath = resolvePath(t, libPath)
	appPath = resolvePath(t, appPath)
	req.LibPath = libPath
	req.AppPath = appPath
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)
	return nil
}
```
