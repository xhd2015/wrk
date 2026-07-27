# Scenario

**Feature**: TTY-simulated confirm accepts treat-as-task for second positional

```
WRK_TASK_LIKE_CONFIRM=1 + stdin y/Y/empty
  wrk <dir> "fix the login bug"
  -> create as --task under WRK_HOME naming
```

## Steps

- Set `ExtraEnv` `WRK_TASK_LIKE_CONFIRM=1` and `StdinInput` for Y/n.
- Leaves set the task-like arg2 text.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
