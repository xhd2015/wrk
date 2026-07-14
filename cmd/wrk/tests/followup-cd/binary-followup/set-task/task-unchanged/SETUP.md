# Scenario

**Feature**: --set-task with same slug writes no follow-up

```
wt with task "my-task"; WRK_FOLLOWUP_FILE set
wrk --set-task "my-task" -> "task unchanged"; follow-up empty
```

## Steps

1. Create worktree with `--task "my-task"`.
2. Run `--set-task "my-task"` with confirm env and follow-up env.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "my-task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.UseFollowupEnv = true
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{"--set-task", "my-task"}
	return nil
}
```
