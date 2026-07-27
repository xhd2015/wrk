# Scenario

**Feature**: diverged branches show Remote: diverged

```
main and origin/main each have unique commits -> Remote: diverged(N commits)
```

## Steps

1. Create tracked repo `{WorkRoot}/div` pushed to bare `origin`.
2. Commit on main locally.
3. Push a different commit to `origin/main` from a clone.
4. Record and run `wrk --projects`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureRemoteBriefHelpersUsed()
	origin := setupRemoteBriefBareOrigin(t, req.WorkRoot, "origin")
	repo := setupRemoteBriefTrackedRepo(t, req.WorkRoot, "div", origin, "diverged base")
	writeFile(t, filepath.Join(repo, "local.txt"), "local\n")
	runGitIsolated(t, repo, "add", "local.txt")
	runGitIsolated(t, repo, "commit", "-m", "local only")
	pushCommitToRemoteBriefOrigin(t, req.WorkRoot, origin, "remote.txt", "remote\n", "remote only")
	runGitIsolated(t, repo, "fetch", "origin")
	recordRemoteBriefProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```