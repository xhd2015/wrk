# Scenario

**Feature**: wrk --list from nested subpath inside main repo

```
# cwd is nested inside main checkout (no .git in subpath)
myrepo/pkg/cmd/tool -> wrk --list -> same output as from repo root
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Create nested subdir `{WorkRoot}/myrepo/pkg/cmd/tool` (no `.git` in subpath).
3. Run `wrk --list` with cwd set to the nested subpath.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)

	subpath := filepath.Join(mainRepo, "pkg", "cmd", "tool")
	mkdirAll(t, subpath)

	req.RepoDir = subpath
	req.Args = []string{"--list"}
	return nil
}
```