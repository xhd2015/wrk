# Scenario

**Feature**: single recorded project with clean main and no linked worktrees

```
wrk --add <repo> -> wrk --projects -> full status block + remote compare + zero worktrees
```

## Steps

1. Create git repo `{WorkRoot}/solo` with upstream tracking `origin/main`.
2. Record via `wrk --add`.
3. Run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "solo", origin, "solo project")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```