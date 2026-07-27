# Scenario

**Feature**: bare `wrk` (no args, non-TTY) enters dashboard — not create

```
# zero positionals and no mode flags
myrepo -> wrk
  -> does NOT create a worktree under WRK_HOME/worktrees
  -> P1 stub: exit 0 or non-zero OK; must not be create path side effects
```

## Steps

- Leaves run wrk with empty Args from a git repo (non-TTY harness).
- Critical assertion: no new path under worktrees from create.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardMainRepo(t, req)
	req.Args = nil
	return nil
}
```
