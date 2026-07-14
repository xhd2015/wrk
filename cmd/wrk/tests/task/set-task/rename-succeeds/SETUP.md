# Scenario

**Feature**: --set-task with TTY confirmation renames worktree and branch

```
# inside linked worktree with task slug, WRK_SET_TASK_CONFIRM=1 bypasses TTY
wrk --set-task "new task" (WRK_SET_TASK_CONFIRM=1)
  -> computes new dir/branch names
  -> git worktree move
  -> git branch -m
  -> prints new path on stdout
```

## Steps

1. Create main repo with initial commit.
2. Spawn consumer linked worktree with `--task "original task"`.
3. Run `wrk --set-task "new task"` from inside the worktree.
4. Verify worktree moved, branch renamed, old paths gone.

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
	req.SetTaskDesc = "new task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
