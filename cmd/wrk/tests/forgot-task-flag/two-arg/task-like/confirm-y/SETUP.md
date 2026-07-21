# Scenario

**Feature**: `--confirm` + y promotes treat-as-task for second positional

```
--confirm + WRK_TASK_LIKE_CONFIRM=1 + stdin y/Y/empty
  wrk <dir> "fix the login bug"
  -> create as --task under WRK_HOME naming
```

## Steps

- Set `--confirm`, `ExtraEnv` `WRK_TASK_LIKE_CONFIRM=1`, and `StdinInput` for Y/n.
- Leaves set the task-like arg2 text.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseScriptTTY = true
		req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
