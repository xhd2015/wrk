# Scenario

**Feature**: user declines treat-as-task → keep second positional as target-dir

```
WRK_TASK_LIKE_CONFIRM=1 + stdin "n\n"
  wrk <dir> <multi-word path under existing parent>
  -> create at that exact target-dir path (current semantics)
```

## Steps

- Confirm env on; stdin `n`.
- Use a moderate multi-word path whose parent exists so target-dir create can succeed.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
