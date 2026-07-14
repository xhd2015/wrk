# Scenario

**Feature**: wrapper --set-task move cds into new worktree path

```
wt with "original task"; source bash.sh; wrk --set-task "new task"
  -> stderr cd <new>; FinalPWD = new path
```

## Steps

1. Create worktree with `--task "original task"`.
2. Run wrapper `--set-task "new task"` with `WRK_SET_TASK_CONFIRM=1`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.StartDir = wtDir
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{"--set-task", "new task"}
	return nil
}
```
