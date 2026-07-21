# Scenario

**Feature**: --set-task on non-TTY auto-yes renames without terminal requirement

```
wrk --set-task "new task desc" (non-TTY) -> exit 0; worktree + branch renamed
```

Default auto-yes skips the rename prompt; --confirm re-enables it and then
requires a TTY (stdout).

## Steps

1. Create a worktree with --task (so branch has wrk naming pattern).
2. Run from inside the worktree via req.RepoDir + req.SetTaskDesc (no WRK_SET_TASK_CONFIRM, no -y).
3. Verify rename succeeds under default auto-yes.

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
	// Force the confirm/TTY path (default auto-yes would skip the prompt).
	req.Args = []string{"--confirm"}
	return nil
}
```
