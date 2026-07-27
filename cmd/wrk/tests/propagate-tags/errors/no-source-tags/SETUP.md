# Scenario

**Feature**: source modules without numeric release tags hard-error on --propagate-tags

```
# lib has go.mod modules but no v*.*.* / sub/v*.*.* tags
cwd=lib -> wrk --propagate-tags --dry-run
  -> exit ≠ 0
  -> stderr indicates missing / no release tags (hard error, not empty plan)
```

## Steps

1. Create lib with root module `example.com/lib` and **no** git tags.
2. Optionally register only lib (or lib+app); run dry-run from lib.
3. Expect hard failure because source has no usable release tags.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	initSingleModuleRepo(t, libPath, "example.com/lib", nil)
	// Intentionally no tagRepo.
	libPath = resolvePath(t, libPath)

	req.SourcePath = libPath
	writeProjectsJSON(t, req.WrkHome, libPath)

	req.RepoDir = libPath
	req.Args = []string{"--propagate-tags", "--dry-run"}
	captureRepoSnapshots(t, req)
	return nil
}
```
