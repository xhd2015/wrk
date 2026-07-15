# Scenario

**Feature**: one-arg task-like non-TTY → error + `wrk -t '…'` hint

```
non-TTY wrk "fix the login bug" (cwd = git repo)
  -> Error: looks like a task description, not a source directory
  -> hint: wrk -t '…'
```

## Steps

- No confirm env; no `-y`.

```go
func Setup(t *testing.T, req *Request) error {
	req.UseScriptTTY = false
	return nil
}
```
