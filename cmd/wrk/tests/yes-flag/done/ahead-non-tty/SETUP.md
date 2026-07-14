# Scenario

**Feature**: `wrk --done -y` merges ahead own worktree on non-TTY without prompt

```
myrepo + wt ahead -> wrk --done -y (non-TTY) -> exit 0; merged + removed
```

## Steps

1. Create linked worktree and commit ahead on it.
2. Run `wrk --done -y` from the worktree (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
