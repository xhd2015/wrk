# Scenario

**Feature**: two-arg spaces + confirm Y → promote to `--task` under WRK_HOME

```
WRK_TASK_LIKE_CONFIRM=1 + stdin "y\n"
  wrk <myrepo> "fix the login bug"
  -> warning + Treat as --task? → create {WRK_HOME}/worktrees/myrepo-main-{date}-fix-the-login-bug
```

## Steps

1. Two-arg with multi-word second positional.
2. Confirm `y`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupTwoArg(t, req, taskLikeSpaces)
	req.StdinInput = "y\n"
	return nil
}
```
