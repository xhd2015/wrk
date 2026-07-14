# Scenario

**Feature**: wrk --done --confirm-from-stdin <wtDir> with ahead commit

```
# wrk creates linked wt; ahead commit on worktree
myrepo (main) + wt (main-{date}) -> ahead commit on wt -> wrk --done <wtDir> --confirm-from-stdin -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk` from main checkout.
2. Commit ahead on worktree.
3. Run `wrk --done --confirm-from-stdin <wtDir>` from WorkRoot with "\n" on stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead work")
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", "--confirm-from-stdin", wtDir}
	req.StdinInput = "\n"
	return nil
}
```