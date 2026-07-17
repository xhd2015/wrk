# Scenario

**Feature**: wrk --done default auto-yes merges ahead branch on non-TTY

```
# wt ahead of main; bare --done (no -y, no --confirm) on non-TTY
myrepo + wt -> commit on wt -> wrk --done -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run bare `wrk --done` without `--confirm` / `-y`.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
