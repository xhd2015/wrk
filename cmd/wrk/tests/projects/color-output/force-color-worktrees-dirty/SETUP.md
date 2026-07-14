# Scenario

**Feature**: --color highlights only the dirty portion of Worktrees summary in red

```
linked worktrees with one dirty -> wrk --projects --color -> red "1 dirty", plain "3 total, "
```

## Steps

1. Create tracked repo `{WorkRoot}/wtdirty`.
2. Add two clean linked worktrees and one dirty linked worktree.
3. Record and run `wrk --projects --color`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "wtdirty", origin, "worktrees dirty color")
	addColorLinkedWorktree(t, repo, "wt-clean-1", "wt-clean-1")
	addColorLinkedWorktree(t, repo, "wt-clean-2", "wt-clean-2")
	dirtyWt := addColorLinkedWorktree(t, repo, "wt-dirty", "wt-dirty")
	dirtyColorWorktree(t, dirtyWt, "dirty.txt", "uncommitted\n")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```