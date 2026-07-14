# Scenario

**Feature**: --color highlights diverged remote summary in red

```
main and origin/main each have unique commits -> wrk --projects --color -> red diverged(N commits)
```

## Steps

1. Create tracked repo `{WorkRoot}/div` pushed to bare `origin`.
2. Commit on main locally.
3. Push a different commit to `origin/main` from a clone.
4. Record and run `wrk --projects --color`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "div", origin, "diverged base")
	writeFile(t, filepath.Join(repo, "local.txt"), "local\n")
	runGitIsolated(t, repo, "add", "local.txt")
	runGitIsolated(t, repo, "commit", "-m", "local only")
	pushCommitToBareOrigin(t, req.WorkRoot, origin, "remote.txt", "remote\n", "remote only")
	runGitIsolated(t, repo, "fetch", "origin")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```