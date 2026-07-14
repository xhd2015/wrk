# Scenario

**Feature**: fetch failure surfaces inline Remote: error on --projects

```
tracked repo with unreachable origin + --fetch -> Remote: error: ... (exit 0)
```

## Steps

1. Create tracked repo with valid origin, then repoint origin to unreachable URL.
2. Record and run `wrk --projects --fetch`.

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "broken-fetch", origin, "broken fetch base")
	runGitIsolated(t, repo, "remote", "set-url", "origin", "file:///nonexistent-wrk-fetch-failure.git")
	recordFetchVerboseProject(t, req, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--projects", "--fetch"}
	return nil
}
```