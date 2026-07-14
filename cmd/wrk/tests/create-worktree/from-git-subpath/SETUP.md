# Scenario

**Feature**: wrk from git subdirectory uses repo root basename

```
# cwd is nested inside main checkout (no .git in subpath)
myrepo/pkg/cmd/tool (main) -> wrk -> {WRK_HOME}/worktrees/myrepo-main-2026-06-30
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Create nested subdir `{WorkRoot}/myrepo/pkg/cmd/tool` (no `.git` in subpath).
3. Set `req.RepoDir` to the nested subpath.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)

	subpath := filepath.Join(mainRepo, "pkg", "cmd", "tool")
	mkdirAll(t, subpath)
	req.RepoDir = subpath
	return nil
}
```