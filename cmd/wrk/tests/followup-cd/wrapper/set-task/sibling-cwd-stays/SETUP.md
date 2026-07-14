# Scenario

**Feature**: wrapper --set-task rename of another worktree while shell cwd is a surviving sibling does not auto-cd

```
# shell cwd = linked wt A; wrapper renames sibling B
source bash.sh from wtA; wrk <wtB> --set-task "new task"
  -> B moved; no stderr cd; FinalPWD stays wtA
```

## Steps

1. Create main; spawn sibling A (`--task "keep here"`) and operated B (`--task "original task"`).
2. Run wrapper `wrk <wtB> --set-task "new task"` with StartDir = wtA and confirm env.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo, "--task", "keep here")
	wtB := runWrkWithArgs(t, req, mainRepo, "--task", "original task")

	req.MainRepo = mainRepo
	req.WtDir = wtB
	req.RepoDir = wtA
	req.StartDir = wtA
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{wtB, "--set-task", "new task"}
	return nil
}
```
