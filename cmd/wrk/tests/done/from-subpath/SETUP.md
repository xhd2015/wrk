# Scenario

**Feature**: wrk --done from nested subpath inside linked worktree

```
# cwd is subdir inside linked wt; resolves checkout root via ShowToplevel
myrepo + wt/pkg/sub -> wrk --done (default auto-yes) -> success
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Create nested subdir inside worktree and set `req.RepoDir` to it.
4. Run `wrk --done` (no confirm flags).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead from subpath")

	subpath := filepath.Join(wtDir, "pkg", "sub")
	mkdirAll(t, subpath)

	req.RepoDir = subpath
	req.Args = []string{"--done"}
	return nil
}
```
