# Scenario

**Feature**: wrk --done <wtDir> with ahead commit (default auto-yes)

```
# wrk creates linked wt; ahead commit on worktree
myrepo (main) + wt (main-{date}) -> ahead commit on wt -> wrk --done <wtDir> -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk` from main checkout.
2. Commit ahead on worktree.
3. Run `wrk --done <wtDir>` from WorkRoot (no confirm flags; default auto-yes).

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead work")
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", wtDir}
	return nil
}
```
