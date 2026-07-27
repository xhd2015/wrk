# Scenario

**Feature**: piped stdout without --color emits plain aligned text

```
tracked clean project -> wrk --projects (pipe) -> no ANSI, Worktrees:    aligned
```

## Steps

1. Create tracked git repo `{WorkRoot}/plain`.
2. Record via `wrk --add`.
3. Run `wrk --projects` (no `--color`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorOutputHelpersUsed()
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "plain", origin, "plain project")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```