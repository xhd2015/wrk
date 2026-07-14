# Scenario

**Feature**: wrk --done rejects non-interactive ahead merge without confirm flag

```
# wt ahead of main; stdin not a TTY and no --confirm-from-stdin
myrepo + wt -> commit on wt -> wrk --done (no stdin) -> error
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done` without `--confirm-from-stdin` and without piped stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```