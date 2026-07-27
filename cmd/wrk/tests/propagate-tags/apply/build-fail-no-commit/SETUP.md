# Scenario

**Feature**: when go build fails after require bump, warn on stderr and do not commit

```
# lib tagged v1.2.3; app requires v1.0.0; app main does not compile
cwd=lib -> wrk --propagate-tags
  -> updates go.mod require to v1.2.3 (tidy ok)
  -> go build ./... fails
  -> stderr: warning: … build …
  -> no commit; app HEAD unchanged
  -> exit 0 (partial success)
  -> stdout: updated block + footer; no go build ok / committed lines
```

## Steps

1. Create tagged `repos/lib` root module `example.com/lib` at `v1.2.3`.
2. Create `repos/app` requiring older `v1.0.0`; write a main that imports lib but
   references a missing symbol so `go build ./...` fails; commit that main.
3. Seed file module proxy for tidy; register both; apply from lib; snapshot first.

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
	// Import kept for tidy; missing symbol → go build ./... fails after apply.
	writeConsumerMainBuildBreak(t, appPath, "example.com/lib")
	runGitIsolated(t, appPath, "add", "main.go")
	runGitIsolated(t, appPath, "commit", "-m", "consumer main broken compile")
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
