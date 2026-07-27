# Scenario

**Feature**: after tidy, successful go build commits consumer go.mod/go.sum with chore(deps) subject

```
# lib tagged v1.2.3; app requires example.com/lib@v1.0.0; main compiles
cwd=lib -> wrk --propagate-tags
  -> updated example.com/app  (project app)
       example.com/lib  v1.0.0 -> v1.2.3
       go build ./... ok
       committed <short7>  chore(deps): bump example.com/lib to v1.2.3
  -> app HEAD advances; commit parent = prior HEAD
  -> commit paths only go.mod / go.sum
  -> go.mod require v1.2.3
```

## Steps

1. Create tagged `repos/lib` root module `example.com/lib` at `v1.2.3`.
2. Create `repos/app` requiring older `v1.0.0`; write a compiling main that imports lib;
   commit the main so the tree is clean before apply.
3. Seed local file module proxy with `example.com/lib@v1.2.3` for offline tidy.
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

	initGitRepo(t, libPath)
	writeGoMod(t, libPath, "example.com/lib", nil)
	writeFile(t, filepath.Join(libPath, "lib.go"), "package lib\n\nfunc Version() string { return \"v1.2.3\" }\n")
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib module")
	tagRepo(t, libPath, "v1.2.3")
	libPath = resolvePath(t, libPath)

	initSingleModuleRepo(t, appPath, "example.com/app", []string{
		"example.com/lib@v1.0.0",
	})
	writeConsumerMainWithImports(t, appPath, "example.com/lib")
	// Commit main so only go.mod/go.sum are dirty after apply (clean post-commit tree).
	runGitIsolated(t, appPath, "add", "main.go")
	runGitIsolated(t, appPath, "commit", "-m", "consumer main imports lib")
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
