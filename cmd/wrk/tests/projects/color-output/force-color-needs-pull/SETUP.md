# Scenario

**Feature**: --color highlights needs pull remote summary in orange

```
main behind origin/main -> wrk --projects --color -> orange needs pull(1 commit behind)
```

## Steps

1. Create tracked repo `{WorkRoot}/behind-main` pushed to bare `origin`.
2. Push an additional commit to `origin/main` from a clone (main stays behind).
3. Record and run `wrk --projects --color`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "behind-main", origin, "behind main base")
	pushCommitToBareOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	runGitIsolated(t, repo, "fetch", "origin")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```