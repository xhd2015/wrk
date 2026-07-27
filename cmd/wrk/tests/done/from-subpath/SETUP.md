# Scenario

**Feature**: wrk --done from nested subpath inside linked worktree

```
# cwd is subdir inside linked wt; resolves checkout root via ShowToplevel
myrepo + wt/pkg/sub -> wrk --done --confirm-from-stdin -> success
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Create nested subdir inside worktree and set `req.RepoDir` to it.
4. Run `wrk --done --confirm-from-stdin` with `\n` on stdin.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead from subpath")

	subpath := filepath.Join(wtDir, "pkg", "sub")
	mkdirAll(t, subpath)

	req.RepoDir = subpath
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```