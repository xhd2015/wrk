# Scenario

**Feature**: wrk --projects counts prunable dead linked worktrees in summary only

```
# linked worktree registered in git but checkout directory deleted
wrk --projects -> exit 0; Worktrees: 0 total, 0 dirty, 1 prune; no per-path prune lines
```

## Steps

1. Create tracked main repo `{WorkRoot}/prune-main`.
2. Add linked worktree `gone-wt`.
3. Delete the `gone-wt` checkout directory (leave git registration).
4. Record and run `wrk --projects` (pipe mode, no `--color`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "prune-main", origin, "prunable worktrees")
	goneWt := addLinkedWorktreeForProject(t, repo, "gone-wt", "gone-wt")
	removeWorktreeCheckout(t, goneWt)
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```