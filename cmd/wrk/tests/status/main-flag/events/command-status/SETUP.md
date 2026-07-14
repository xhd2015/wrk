# Scenario

**Feature**: events.jsonl command is "status" for wrk --main --status

```
external wt -> wrk --main --status
  -> exit 0
  -> last event: command=status, args=[--main,--status], exit_code=0
```

## Steps

1. Create main + external wrk worktree; cwd = external.
2. Args = `--main`, `--status`.
3. Assert last event only (no second status invocation before read).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, branch := setupExternalMainFlagFixture(t, req)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch
	req.RepoDir = wtDir
	setMainStatusArgs(req, "--main", "--status")
	return nil
}
```