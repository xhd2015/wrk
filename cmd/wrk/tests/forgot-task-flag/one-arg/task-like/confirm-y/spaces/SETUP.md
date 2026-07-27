# Scenario

**Feature**: one-arg spaces + confirm Y → promote create from current directory

```
WRK_TASK_LIKE_CONFIRM=1 + "y\n"
  wrk "fix the login bug"  # cwd=myrepo
  -> worktree myrepo-main-{date}-fix-the-login-bug under WRK_HOME
```

## Steps

1. setupOneArg with multi-word text.
2. Stdin `y`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupOneArg(t, req, taskLikeSpaces)
	req.StdinInput = "y\n"
	return nil
}
```
