# Scenario

**Feature**: Remote shows branch ahead of upstream

```
main tracked to origin/main -> commit on main -> Remote: needs push(+1 commit)
```

## Steps

1. Create tracked repo `{WorkRoot}/ahead` pushed to bare `origin`.
2. Commit once more on `main` (1 commit ahead of `origin/main`).
3. Record and run `wrk --projects`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "ahead", origin, "ahead base")
	writeFile(t, filepath.Join(repo, "ahead.txt"), "ahead\n")
	runGitIsolated(t, repo, "add", "ahead.txt")
	runGitIsolated(t, repo, "commit", "-m", "ahead of upstream")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```