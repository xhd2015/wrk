# Scenario

**Feature**: wrk --done when branch is already included in main

```
# wrk creates linked wt; commit merged into main; wt still exists at same commits
myrepo (main) + wt (main-{date}) -> merge wt branch into main -> wrk --done -> remove only
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree and fast-forward merge branch into main.
3. Run `wrk --done` from worktree root (no confirmation needed).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```