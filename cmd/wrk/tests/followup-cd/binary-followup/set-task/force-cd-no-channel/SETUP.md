# Scenario

**Feature**: --set-task rename with --force-cd from sibling, channel closed, launches shell at newPath (Branch B)

```
wtA (cwd) + wtB task "original task"; WRK_FOLLOWUP_FILE unset; fake bash
wrk <wtB> --set-task "new task" --force-cd
  -> stdout new path (set-task contract)
  -> stderr install hint
  -> fake shell cwd = newPath
```

## Steps

1. Create sibling A and operated B; install fake bash.
2. Run set-task with `--force-cd`, confirm env, no follow-up env; cwd = wtA.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo, "--task", "keep here")
	wtB := runWrkWithArgs(t, req, mainRepo, "--task", "original task")

	req.MainRepo = mainRepo
	req.WtDir = wtB
	req.RepoDir = wtA
	req.UseFollowupEnv = false
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.CLIArgs = []string{wtB, "--set-task", "new task", "--force-cd"}
	installFakeBash(t, req, 0)
	return nil
}
```
