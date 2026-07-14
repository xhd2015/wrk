# Scenario

**Feature**: wrk --projects -v on clean tracked repo has empty stderr

```
recorded tracked repo -> wrk --projects -v -> stderr empty
```

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "proj-v-clean", origin, "proj v clean")
	recordFetchVerboseProject(t, req, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--projects", "-v"}
	return nil
}
```