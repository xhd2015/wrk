# Scenario

**Feature**: non-TTY task-like second positional → hard error + `-t` hint (no promote)

```
pipe stdin (non-TTY) + wrk <dir> <task-like>
  -> exit != 0; stderr looks like task / not target directory; hint -t
  -> no worktree created
```

## Steps

- Leave `UseScriptTTY` false; do not set `WRK_TASK_LIKE_CONFIRM`.
- Leaves vary task-like reason (spaces / >120 / >255).

```go
func Setup(t *testing.T, req *Request) error {
	req.UseScriptTTY = false
	return nil
}
```
