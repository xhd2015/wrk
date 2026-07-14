# Scenario

**Feature**: --color highlights dirty word and non-zero count segments in red

```
main repo with 1 added file -> wrk --projects --color -> red dirty + 1 added, grey zero segments
```

## Steps

1. Create tracked git repo `{WorkRoot}/dirty-main`.
2. Add uncommitted file on main (1 added).
3. Record and run `wrk --projects --color`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "dirty-main", origin, "dirty main base")
	writeFile(t, filepath.Join(repo, "untracked.txt"), "dirty\n")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```