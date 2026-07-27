# Scenario

**Feature**: second positional is task-like (spaces, long, or ENAMETOOLONG-class)

```
wrk <dir> <task-like> -> treat-as-task decision (TTY / WRK_TASK_LIKE_CONFIRM / -y / non-TTY)
```

## Steps

- Descendants choose interactive vs non-TTY vs `-y` and the task-like reason.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping node: leaves configure positionals and interactive mode.
	skipIfNoGit(t)
	return nil
}
```
