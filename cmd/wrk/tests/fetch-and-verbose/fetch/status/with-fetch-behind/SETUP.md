# Scenario

**Feature**: wrk --status --fetch from main root reveals behind-upstream

```
stale fixture + --status --fetch -> root Remote: needs pull(1 commit behind)
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "status-fetch", origin, "status fetch base")
	pushCommitToFetchVerboseOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--status", "--fetch"}
	return nil
}
```