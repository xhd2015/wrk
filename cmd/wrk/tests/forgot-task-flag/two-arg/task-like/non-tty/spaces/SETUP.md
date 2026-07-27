# Scenario

**Feature**: two-arg multi-word second positional, non-TTY → error + hint

```
wrk <myrepo> "fix the login bug" (non-TTY)
  -> Error: looks like a task description, not a target directory
  -> hint: … -t …
```

## Steps

1. Init `myrepo` on main.
2. Run `wrk <myrepo> "fix the login bug"` from WorkRoot without confirm env.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoArg(t, req, taskLikeSpaces)
	return nil
}
```
