# Scenario

**Feature**: `wrk --merge-back -y` merges ahead worktree on non-TTY, keeps worktree

```
wt ahead -> wrk --merge-back -y (non-TTY) -> exit 0; merged, wt kept
```

## Steps

1. Create linked worktree and commit ahead on it.
2. Run `wrk --merge-back -y` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--merge-back", "-y"}
	return nil
}
```
