# Scenario

**Feature**: --color highlights only non-zero dirty count segments in red

```
main repo with 2 modified tracked files -> wrk --projects --color -> grey zero segments, red changed segment
```

## Steps

1. Create tracked git repo `{WorkRoot}/partial-dirty`.
2. Commit two tracked files, then modify both (0 staged, 2 changed).
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
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "partial-dirty", origin, "partial dirty base")
	writeFile(t, filepath.Join(repo, "one.txt"), "one v1\n")
	writeFile(t, filepath.Join(repo, "two.txt"), "two v1\n")
	runGitIsolated(t, repo, "add", "one.txt", "two.txt")
	runGitIsolated(t, repo, "commit", "-m", "add tracked files")
	writeFile(t, filepath.Join(repo, "one.txt"), "one v2\n")
	writeFile(t, filepath.Join(repo, "two.txt"), "two v2\n")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```