# Scenario

**Feature**: consumer already at source release versions → no would-update module blocks

```
# lib @ v1.2.3; app requires example.com/lib@v1.2.3 (same)
cwd=lib -> wrk --propagate-tags --dry-run
  -> source: block shown
  -> no "would: update example.com/app" block
  -> footer would: update 0 modules across 0 projects
  -> no mutation
```

## Steps

1. Create tagged lib `example.com/lib` at `v1.2.3`.
2. Create app requiring the **same** version `v1.2.3`.
3. Register both; dry-run from lib.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	initSingleModuleRepo(t, libPath, "example.com/lib", nil)
	tagRepo(t, libPath, "v1.2.3")
	libPath = resolvePath(t, libPath)

	initSingleModuleRepo(t, appPath, "example.com/app", []string{
		"example.com/lib@v1.2.3",
	})
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
