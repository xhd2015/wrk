# Scenario

**Feature**: wrk --projects --fetch -v logs fetch to stderr

```
tracked repo with upstream -> --projects --fetch -v -> stderr fetch line
```

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "proj-v-fetch", origin, "proj v fetch")
	recordFetchVerboseProject(t, req, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--projects", "--fetch", "-v"}
	return nil
}
```