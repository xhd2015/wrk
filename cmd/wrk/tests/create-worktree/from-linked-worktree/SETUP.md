# Scenario

**Feature**: wrk from linked worktree uses main-repo basename

```
# main repo myrepo + linked worktree; cwd is linked checkout
linked worktree cwd -> wrk -> {WRK_HOME}/worktrees/myrepo-linked-side-2026-06-30
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `{WorkRoot}/linked-wt` on branch `linked-side`.
3. Set `req.RepoDir` to the linked worktree path.

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

	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "linked-side", linkedWT)
	req.RepoDir = linkedWT
	return nil
}
```