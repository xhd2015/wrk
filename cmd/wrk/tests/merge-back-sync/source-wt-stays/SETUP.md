# Scenario

**Feature**: `--merge-back --sync` keeps source worktree on disk after success

```
# single ahead wt; merge-back keeps it; zero-summary sync from main
myrepo + wtA (ahead)
  -> wrk --merge-back -y --sync
  -> merged branch <wtA> into main
  -> <blank>
  -> synced: 0 into main, 0 into worktrees, 0 skipped
  -> wtA still on disk; branch kept; no "worktree removed:"
```

## Steps

1. Create wrk-managed linked worktree and commit ahead.
2. Run `wrk --merge-back -y --sync` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--merge-back", "-y", "--sync"}
	return nil
}
```
