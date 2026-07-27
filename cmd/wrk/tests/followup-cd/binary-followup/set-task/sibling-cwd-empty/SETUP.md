# Scenario

**Feature**: --set-task rename of another worktree while shell cwd is a surviving sibling writes no follow-up

```
# shell cwd = linked wt A (still valid after); rename sibling B
wtA (cwd) + wtB task "original task" + WRK_FOLLOWUP_FILE + WRK_SET_TASK_CONFIRM=1
wrk <wtB> --set-task "new task" -> B moved; follow-up empty (A still exists)
```

## Steps

1. Create main repo; spawn sibling wt A (`--task "keep here"`) and operated wt B (`--task "original task"`).
2. Run `wrk <wtB> --set-task "new task"` with process cwd = wtA, confirm + follow-up env set.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo, "--task", "keep here")
	wtB := runWrkWithArgs(t, req, mainRepo, "--task", "original task")

	req.MainRepo = mainRepo
	req.WtDir = wtB
	// Process/shell cwd remains sibling A (not the operated tree).
	req.RepoDir = wtA
	req.UseFollowupEnv = true
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{wtB, "--set-task", "new task"}
	return nil
}
```
