# Scenario

**Feature**: main matches upstream shows Remote: identical

```
main tracked to origin/main at same commit -> Remote: identical
```

## Steps

1. Create tracked repo `{WorkRoot}/synced` pushed to bare `origin`.
2. Record and run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureRemoteBriefHelpersUsed()
	origin := setupRemoteBriefBareOrigin(t, req.WorkRoot, "origin")
	repo := setupRemoteBriefTrackedRepo(t, req.WorkRoot, "synced", origin, "synced base")
	recordRemoteBriefProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```