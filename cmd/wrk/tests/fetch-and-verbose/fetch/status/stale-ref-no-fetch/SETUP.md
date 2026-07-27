# Scenario

**Feature**: wrk --status without --fetch uses stale tracking ref on root Remote:

```
push to origin without local fetch -> wrk --status -> root Remote: identical
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "status-stale", origin, "status stale base")
	pushCommitToFetchVerboseOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--status"}
	return nil
}
```