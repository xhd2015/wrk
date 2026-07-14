# Scenario

**Feature**: wrk --done ff-merges ahead branch with piped confirmation

```
# wt branch ahead of main; user confirms via --confirm-from-stdin + Enter
myrepo + wt -> commit on wt -> wrk --done --confirm-from-stdin (\n) -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done --confirm-from-stdin` with `\n` on stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```