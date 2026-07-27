# Scenario

**Feature**: wrk --projects prints minimal block when a recorded main repo is broken

```
# project recorded while healthy; .git removed before --projects
wrk --projects -> exit 0; block is only Dir + Status error (no Branch/Commit/Remote/Worktrees)
```

## Steps

1. Create tracked main repo `{WorkRoot}/broken-main`.
2. Record via `wrk --add`.
3. Delete `{WorkRoot}/broken-main/.git` so the path is no longer a git repo.
4. Run `wrk --projects` (pipe mode, no `--color`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "broken-main", origin, "broken main repo")
	recordProject(t, req, repo)
	removeGitDir(t, repo)
	req.MainRepo = repo
	return nil
}
```