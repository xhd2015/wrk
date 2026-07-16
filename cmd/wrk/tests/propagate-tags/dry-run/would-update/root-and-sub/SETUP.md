# Scenario

**Feature**: source root + nested sub releases plan would-update for outdated consumer

```
# lib: example.com/lib @ v1.2.3 (tag v1.2.3)
#      example.com/lib/sub @ v0.1.0 (tag sub/v0.1.0)
# app: requires lib@v1.0.0 and lib/sub@v0.0.1
repos registered: lib, app
cwd=lib -> wrk --propagate-tags --dry-run
  -> source: block with both releases
  -> would: update example.com/app with two arrows
  -> footer would: update 2 modules across 1 project
  -> go.mod / tags / HEAD unchanged
```

## Steps

1. Create `repos/lib` with root + `sub/` modules; tag `v1.2.3` and `sub/v0.1.0`.
2. Create `repos/app` requiring both modules at older versions.
3. Register both paths; run dry-run from `lib`.
4. Capture pre-run go.mod / HEAD / tags snapshots.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	initRootAndSubModuleRepo(t, libPath, "example.com/lib")
	tagRepo(t, libPath, "v1.2.3")
	tagRepo(t, libPath, "sub/v0.1.0")

	initSingleModuleRepo(t, appPath, "example.com/app", []string{
		"example.com/lib@v1.0.0",
		"example.com/lib/sub@v0.0.1",
	})

	libPath = resolvePath(t, libPath)
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
