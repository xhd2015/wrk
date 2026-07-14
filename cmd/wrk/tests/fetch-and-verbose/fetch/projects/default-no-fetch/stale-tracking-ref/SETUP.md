# Scenario

**Feature**: stale origin/main tracking ref yields Remote: identical without --fetch

```
tracked repo; push commit to origin; no fetch in main -> wrk --projects -> identical
```

## Steps

1. Create tracked repo pushed to bare origin.
2. Push additional commit to origin from clone (main stays behind, no fetch).
3. Record project and run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "stale-main", origin, "stale base")
	pushCommitToFetchVerboseOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	recordFetchVerboseProject(t, req, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--projects"}
	return nil
}
```