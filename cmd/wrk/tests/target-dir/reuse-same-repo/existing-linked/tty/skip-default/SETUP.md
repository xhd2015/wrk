# Scenario

**Feature**: TTY named bring with existing linked WT auto-skips by default (no prompt)

```
# existing WRK_HOME worktree of myrepo
# default auto-skip (Y default) without --confirm
# stdout = existing path; no new under target
```

## Steps

1. Pre-create one linked worktree of `myrepo` under `WRK_HOME/worktrees`.
2. Run named bring under fake TTY (no stdin answer required).

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	return nil
}
```
