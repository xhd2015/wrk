# Scenario

**Feature**: --status --fetch -v from main root logs fetch to stderr

```
tracked main repo + --status --fetch -v -> stderr contains fetch log line
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "status-fetch-v", origin, "status fetch v base")
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--status", "--fetch", "-v"}
	return nil
}
```