# Scenario

**Feature**: wrk --done of already-included worktree succeeds even when main is dirty

```
# origin present so main-sync would otherwise require a clean main
myrepo (dirty main, origin) + wt already merged into main
  -> wrk --done
  -> worktree removed; main stays dirty
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree and fast-forward merge branch into main.
3. Add `origin` and push so remote sync is eligible.
4. Leave an uncommitted file on main.
5. Run `wrk --done` from the worktree.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	bare := filepath.Join(req.WorkRoot, "origin.git")
	runGitIsolated(t, req.WorkRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")

	writeFile(t, filepath.Join(mainRepo, "dirty-main.txt"), "uncommitted")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
