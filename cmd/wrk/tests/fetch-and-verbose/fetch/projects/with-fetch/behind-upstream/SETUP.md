# Scenario

**Feature**: --fetch reveals behind-upstream state after origin push

```
stale tracking ref fixture + wrk --projects --fetch -> needs pull(1 commit behind)
```

## Steps

1. Create tracked repo; push extra commit to origin without local fetch.
2. Record and run `wrk --projects --fetch`.

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "fetch-behind", origin, "fetch behind base")
	pushCommitToFetchVerboseOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	recordFetchVerboseProject(t, req, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--projects", "--fetch"}
	return nil
}
```