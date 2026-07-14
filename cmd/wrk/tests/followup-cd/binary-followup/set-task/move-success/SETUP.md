# Scenario

**Feature**: --set-task move success writes cd to new path

```
wt with task "original task"; WRK_SET_TASK_CONFIRM=1 + WRK_FOLLOWUP_FILE
wrk --set-task "new task" -> follow-up: cd <newPath>
```

## Steps

1. Create main repo and worktree with `--task "original task"`.
2. Run `--set-task "new task"` with confirm env and follow-up env.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.UseFollowupEnv = true
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{"--set-task", "new task"}
	return nil
}
```
