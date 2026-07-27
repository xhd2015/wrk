# Scenario

**Feature**: non-TTY named bring refuses when a live linked worktree of source already exists

```
# existing WRK_HOME worktree + non-TTY
# stderr: wrk: <basename> already has a linked worktree at <path>; refusing non-interactive create (default is skip; re-run in a TTY)
# empty stdout; no new worktree under target
```

## Steps

1. Pre-create one linked worktree of `myrepo`.
2. Run `wrk myrepo {WorkRoot}/target` without TTY (default doctest pipe).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	return nil
}
```
