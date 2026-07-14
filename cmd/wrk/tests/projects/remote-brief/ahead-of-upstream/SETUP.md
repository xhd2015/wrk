# Scenario

**Feature**: main ahead of upstream shows Remote: needs push

```
main tracked to origin/main -> one local commit -> Remote: needs push(+1 commit)
```

## Steps

1. Create tracked repo `{WorkRoot}/ahead` pushed to bare `origin`.
2. Commit once more on `main` (1 commit ahead).
3. Record and run `wrk --projects`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureRemoteBriefHelpersUsed()
	origin := setupRemoteBriefBareOrigin(t, req.WorkRoot, "origin")
	repo := setupRemoteBriefTrackedRepo(t, req.WorkRoot, "ahead", origin, "ahead base")
	writeFile(t, filepath.Join(repo, "ahead.txt"), "ahead\n")
	runGitIsolated(t, repo, "add", "ahead.txt")
	runGitIsolated(t, repo, "commit", "-m", "ahead of upstream")
	recordRemoteBriefProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```