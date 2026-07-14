# Scenario

**Feature**: `wrk --done -y` on TTY shows no `Proceed?` prompt

```
wt ahead + fake TTY -> wrk --done -y -> success without Proceed? in output
```

## Steps

1. Create linked worktree and commit ahead on it.
2. Run `wrk --done -y` under `script` fake TTY.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--done", "-y"}
	req.UseScriptTTY = true
	return nil
}
```
