# Scenario

**Feature**: wrk --done aborts when user declines confirmation

```
# wt ahead of main; --confirm restores prompt; user types 'n'
myrepo + wt -> commit on wt -> wrk --done --confirm --confirm-from-stdin (n) -> merge-back aborted
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done --confirm --confirm-from-stdin` with `n\n` on stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm", "--confirm-from-stdin"}
	req.StdinInput = "n\n"
	return nil
}
```