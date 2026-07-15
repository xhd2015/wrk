# Scenario

**Feature**: one-arg task-like + confirm Y → create from cwd with task

```
WRK_TASK_LIKE_CONFIRM=1 + stdin y
  (cd myrepo && wrk "fix the login bug")
  -> create WRK_HOME worktree with slug from task
```

## Steps

- Confirm env; leaves set stdin + arg1.

```go
func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
