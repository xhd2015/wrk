# Scenario

**Feature**: --set-task rename of sibling with --force-cd writes follow-up to newPath (cwd gate bypass)

```
# shell cwd = linked wt A (still valid after); rename sibling B
wtA (cwd) + wtB task "original task" + WRK_FOLLOWUP_FILE + WRK_SET_TASK_CONFIRM=1
wrk <wtB> --set-task "new task" --force-cd -> B moved; follow-up: cd <newPath-abs>
```

## Steps

1. Create main repo; spawn sibling wt A (`--task "keep here"`) and operated wt B (`--task "original task"`).
2. Run `wrk <wtB> --set-task "new task" --force-cd` with process cwd = wtA, confirm + follow-up env set.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo, "--task", "keep here")
	wtB := runWrkWithArgs(t, req, mainRepo, "--task", "original task")

	req.MainRepo = mainRepo
	req.WtDir = wtB
	// Process/shell cwd remains sibling A (cwd-missing gate would normally skip).
	req.RepoDir = wtA
	req.UseFollowupEnv = true
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{wtB, "--set-task", "new task", "--force-cd"}
	return nil
}
```
