# Scenario

**Feature**: wrk --projects reports stale-gitdir linked worktrees as errors without aborting

```
# main repo healthy; one linked wt ok, one linked wt with stale .git gitdir
wrk --projects -> exit 0; Worktrees summary counts error; per-path detail line
```

Reproduces: `fatal: not a git repository: .../old-main/.git/worktrees/<name>` when the
checkout directory exists but its `.git` file still references the pre-move main repo.

## Steps

1. Create tracked main repo `{WorkRoot}/new-main`.
2. Add linked worktree `good-wt` (healthy).
3. Add linked worktree `stale-wt`, then overwrite its `.git` with a stale `gitdir:` path under `{WorkRoot}/old-main/.git/worktrees/...` (old main never exists).
4. Record and run `wrk --projects` (pipe mode, no `--color`).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "new-main", origin, "main after move")
	addLinkedWorktreeForProject(t, repo, "good-wt", "good-wt")
	staleWt := addLinkedWorktreeForProject(t, repo, "stale-wt", "stale-wt")
	staleGitdir := filepath.Join(req.WorkRoot, "old-main", ".git", "worktrees", "stale-wt")
	writeFile(t, filepath.Join(staleWt, ".git"), "gitdir: "+staleGitdir+"\n")
	recordProject(t, req, repo)
	req.MainRepo = repo
	req.WtDir = staleWt
	return nil
}
```