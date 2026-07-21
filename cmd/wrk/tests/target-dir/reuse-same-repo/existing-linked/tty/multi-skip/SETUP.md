# Scenario

**Feature**: multi linked WTs + default auto-skip reuses lex-smallest

```
# two WRK_HOME worktrees of myrepo
# default auto-skip -> stdout = lex-smallest path; no new under target
```

## Steps

1. Pre-create two linked worktrees of `myrepo`.
2. Run named bring under fake TTY without `--confirm`.

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 2)
	req.WtDir = paths[0]
	req.ExternalWtDir2 = paths[1]
	return nil
}
```
