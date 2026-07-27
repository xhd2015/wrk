# Scenario

**Feature**: apply drops cross-project local replace, bumps require, builds, commits

```
# lib tagged v1.2.3
# app requires example.com/lib@v1.0.0 + replace example.com/lib => <abs-lib>
cwd=lib -> wrk --propagate-tags
  -> updated example.com/app  (lib v1.0.0 -> v1.2.3)
       go build ./... ok
       committed <short7>  chore(deps): bump example.com/lib to v1.2.3
  -> dropped replace example.com/lib  (project app)
  -> app go.mod: require v1.2.3, no replace directive; HEAD advances
```

## Steps

1. Create tagged `repos/lib` root module `example.com/lib` at `v1.2.3`.
2. Create `repos/app` with older require **and** local replace pointing at lib.
3. Consumer main imports the lib package; seed file module proxy for tidy after drop.
4. Register both; run apply from lib; snapshot go.mod/tags/HEAD first.

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

	// Library module: single package "lib" (no package main) so consumers can import it.
	initGitRepo(t, libPath)
	writeGoMod(t, libPath, "example.com/lib", nil)
	writeFile(t, filepath.Join(libPath, "lib.go"), "package lib\n\nfunc Version() string { return \"v1.2.3\" }\n")
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib module")
	tagRepo(t, libPath, "v1.2.3")
	libPath = resolvePath(t, libPath)

	// Local replace to absolute lib path (cross-project).
	initSingleModuleRepo(t, appPath, "example.com/app",
		[]string{"example.com/lib@v1.0.0"},
		"example.com/lib=>"+libPath,
	)
	writeConsumerMainWithImports(t, appPath, "example.com/lib")
	appPath = resolvePath(t, appPath)

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, "example.com/lib", "v1.2.3", libPath)
	enableFileModuleProxy(t, req, proxyRoot)

	req.SourcePath = libPath
	req.AppPath = appPath
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)

	req.RepoDir = libPath
	req.Args = []string{"--propagate-tags"}
	captureRepoSnapshots(t, req)
	return nil
}
```
