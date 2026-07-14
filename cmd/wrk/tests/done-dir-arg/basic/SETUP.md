# Scenario

**Feature**: wrk --done <wtDir> when branch is already in main

```
# wrk creates linked wt; commit merged into main; wt still exists at same commits
myrepo (main) + wt (main-{date}) -> merge wt branch into main -> wrk --done <wtDir> -> remove only
```

## Steps

1. Create main repo and linked worktree via `wrk` from main checkout.
2. Commit on worktree and fast-forward merge branch into main.
3. Run `wrk --done <wtDir>` from WorkRoot (not from inside the worktree).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", wtDir}
	return nil
}
```