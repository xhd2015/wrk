# Scenario

**Feature**: main behind upstream shows Remote: needs pull with commit count

```
main tracked to origin/main -> origin has extra commit -> Remote: needs pull(1 commit behind)
```

## Steps

1. Create tracked repo `{WorkRoot}/behind` pushed to bare `origin`.
2. Push an additional commit to `origin/main` from a clone (main stays behind).
3. Record and run `wrk --projects`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureRemoteBriefHelpersUsed()
	origin := setupRemoteBriefBareOrigin(t, req.WorkRoot, "origin")
	repo := setupRemoteBriefTrackedRepo(t, req.WorkRoot, "behind", origin, "behind base")
	pushCommitToRemoteBriefOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	runGitIsolated(t, repo, "fetch", "origin")
	recordRemoteBriefProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```