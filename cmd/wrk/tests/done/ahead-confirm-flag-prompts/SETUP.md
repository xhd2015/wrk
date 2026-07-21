# Scenario

**Feature**: `--confirm` re-enables Proceed? prompt; default Enter accepts

```
# wt ahead; --confirm + --confirm-from-stdin + Enter → shows Proceed? and merges
myrepo + wt ahead -> wrk --done --confirm --confirm-from-stdin (\n) -> Proceed? then ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run with `--confirm --confirm-from-stdin` and stdin `\n` (default Y).

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```
