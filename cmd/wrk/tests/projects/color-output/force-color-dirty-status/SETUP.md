# Scenario

**Feature**: --color highlights dirty word; added segments green; other non-zero red

```
main repo with 1 staged added file -> wrk --projects --color -> red dirty + green 1 staged, grey zero segments
```

## Steps

1. Create tracked git repo `{WorkRoot}/dirty-main`.
2. Stage a new file on main (1 staged).
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
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "dirty-main", origin, "dirty main base")
	writeFile(t, filepath.Join(repo, "staged.txt"), "dirty\n")
	runGitIsolated(t, repo, "add", "staged.txt")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```
