# Scenario

**Feature**: --color highlights worktree error summary and detail segments in red

```
# same stale-gitdir setup as detailed-status/stale-gitdir-linked
wrk --projects --color -> red on "1 error" and per-path "error: ..." values
```

## Steps

1. Create tracked main repo `{WorkRoot}/new-main`.
2. Add linked worktree `good-wt` (healthy).
3. Add linked worktree `stale-wt`, overwrite `.git` with stale gitdir under non-existent `old-main`.
4. Record and run `wrk --projects --color`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "new-main", origin, "main after move")
	addColorLinkedWorktree(t, repo, "good-wt", "good-wt")
	staleWt := addColorLinkedWorktree(t, repo, "stale-wt", "stale-wt")
	staleGitdir := filepath.Join(req.WorkRoot, "old-main", ".git", "worktrees", "stale-wt")
	writeFile(t, filepath.Join(staleWt, ".git"), "gitdir: "+staleGitdir+"\n")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	req.WtDir = staleWt
	return nil
}
```