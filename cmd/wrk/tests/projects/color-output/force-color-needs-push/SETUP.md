# Scenario

**Feature**: --color highlights needs push remote summary in orange

```
main ahead of origin/main -> wrk --projects --color -> orange needs push(...)
```

## Steps

1. Create tracked repo `{WorkRoot}/push` pushed to bare `origin`.
2. Commit once more on `main` (1 commit ahead).
3. Record and run `wrk --projects --color`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "push", origin, "push base")
	writeFile(t, filepath.Join(repo, "ahead.txt"), "ahead\n")
	runGitIsolated(t, repo, "add", "ahead.txt")
	runGitIsolated(t, repo, "commit", "-m", "ahead of upstream")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```