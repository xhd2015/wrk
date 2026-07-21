# Scenario

**Feature**: one-arg task-like + `--confirm` + y → create from cwd with task

```
--confirm + WRK_TASK_LIKE_CONFIRM=1 + stdin y
  (cd myrepo && wrk "fix the login bug")
  -> create WRK_HOME worktree with slug from task
```

## Steps

- `--confirm` + confirm env; leaves set stdin + arg1.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseScriptTTY = true
		req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
