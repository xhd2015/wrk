# Scenario

**Feature**: wrk --done aborts when user declines confirmation

```
# wt ahead of main; user types 'n' via --confirm-from-stdin
myrepo + wt -> commit on wt -> wrk --done --confirm-from-stdin (n) -> merge-back aborted
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done --confirm-from-stdin` with `n\n` on stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "n\n"
	return nil
}
```