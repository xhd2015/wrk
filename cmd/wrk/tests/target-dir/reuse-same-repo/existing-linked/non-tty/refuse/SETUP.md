# Scenario

**Feature**: non-TTY named bring refuses when a live linked worktree of source already exists

```
# existing WRK_HOME worktree + non-TTY
# stderr: wrk: <basename> already exists in <path>; refusing non-interactive create …
# empty stdout; no new worktree under target
```

## Steps

1. Pre-create one linked worktree of `myrepo`.
2. Run `wrk myrepo {WorkRoot}/target` without TTY (default doctest pipe).

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	return nil
}
```
