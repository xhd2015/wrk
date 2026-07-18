# Scenario

**Feature**: wrk -v fixed-path create with pre-existing branch uses always-new `-b` (P0 C5 behavior-change)

```
# branch main-{date} pre-exists; fixed <target-dir> creates NEW branch main-{date}-1 via worktree add -b
# never reuses via --no-checkout
myrepo (main) + refs/heads/main-2026-06-30 -> wrk myrepo <wt> -v
  -> path <wt>, branch main-2026-06-30-1; stderr streams worktree add -b (not --no-checkout)
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initFetchVerboseRepo(t, repo, "create v branch collision always-new")
	runGitIsolated(t, repo, "branch", branchName("main", wrkDate, 0))
	req.TargetDir = repo
	req.RepoDir = req.WorkRoot
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"--new", "-v"}
	return nil
}
```
