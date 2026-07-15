# Scenario

**Feature**: `--done --sync` with no remaining linked worktrees prints zero-summary sync

```
# only wtA; after done, main-only → zero-action sync after blank line
myrepo + wtA (ahead)
  -> wrk --done -y --sync
  -> merged branch <wtA> into main
  -> <blank>
  -> synced: 0 into main, 0 into worktrees, 0 skipped
```

## Steps

1. Create wrk-managed linked worktree and commit ahead.
2. Run `wrk --done -y --sync` from the worktree (no second worktree).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--done", "-y", "--sync"}
	return nil
}
```
