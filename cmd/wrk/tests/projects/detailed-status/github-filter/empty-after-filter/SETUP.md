# Scenario

**Feature**: only non-github origins → empty list

```
local bare origin only -> wrk --projects --github -> exit 0, empty stdout
```

## Steps

1. Record one repo with local bare origin.
2. Run `wrk --projects --github`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureGitHubFilterHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin-only-local")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "local-only", origin, "local only")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```
