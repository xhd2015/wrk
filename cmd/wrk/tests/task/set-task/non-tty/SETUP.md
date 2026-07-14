# Scenario

**Feature**: --set-task in non-TTY environment errors with "requires terminal"

```
wrk --set-task "new desc" (non-TTY) -> non-zero exit, stderr "requires terminal"
```

## Steps

1. Create a worktree with --task (so branch has wrk naming pattern).
2. Run from inside the worktree via req.RepoDir + req.SetTaskDesc.
3. Verify non-zero exit and error about terminal.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	// Create worktree with --task
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "new task desc"
	return nil
}
```