# Scenario

**Feature**: project block reports linked worktree clean/dirty counts

```
main repo + 2 clean linked wts + 1 dirty linked wt -> Worktrees: 3 total, 1 dirty
```

## Steps

1. Create tracked git repo `{WorkRoot}/mixed`.
2. Add linked worktrees `wt-clean-1`, `wt-clean-2`, `wt-dirty`.
3. Leave an uncommitted change in `wt-dirty`.
4. Record and run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "mixed", origin, "mixed worktrees")
	addLinkedWorktreeForProject(t, repo, "wt-clean-1", "wt-clean-1")
	addLinkedWorktreeForProject(t, repo, "wt-clean-2", "wt-clean-2")
	dirtyWt := addLinkedWorktreeForProject(t, repo, "wt-dirty", "wt-dirty")
	dirtyWorktree(t, dirtyWt, "dirty.txt", "uncommitted\n")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```