# Scenario

**Feature**: one-arg spaces + `-y` → create from cwd with task

```
wrk "fix the login bug" -y   # cwd=myrepo
  -> WRK_HOME worktree + slug
```

## Steps

1. setupOneArg spaces; parent adds `-y`.

```go
func Setup(t *testing.T, req *Request) error {
	setupOneArg(t, req, taskLikeSpaces)
	return nil
}
```
