# Scenario

**Feature**: wrk --done auto-yes merges ahead own worktree on non-TTY (no confirm flag)

```
# wt ahead of main; stdin not a TTY; default auto-yes (no --confirm)
myrepo + wt -> commit on wt -> wrk --done -> exit 0; ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done` without `--confirm` and without piped stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
