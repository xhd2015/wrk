# Scenario

**Feature**: user declines treat-as-task via `--confirm` + `n` → keep second positional as target-dir

```
--confirm + WRK_TASK_LIKE_CONFIRM=1 + stdin "n\n"
  wrk <dir> <multi-word path under existing parent>
  -> create at that exact target-dir path (current semantics)
```

## Steps

- `--confirm` forces interactive treat-as-task prompt; confirm env for non-TTY stdin; stdin `n`.
- Use a moderate multi-word path whose parent exists so target-dir create can succeed.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseScriptTTY = true
		req.ExtraEnv = append(req.ExtraEnv, envTaskLikeConfirm)
	return nil
}
```
