# Scenario

**Feature**: TTY named bring with existing linked WT + default Enter skips create

```
# existing WRK_HOME worktree of myrepo
# TTY prompt + Enter (default Y) -> skip; stdout = existing path; no new under target
myrepo-main-{date} exists
  -> wrk myrepo {WorkRoot}/target  (stdin \n)
  -> stdout existing path; no target/myrepo-main-{date}
```

## Steps

1. Pre-create one linked worktree of `myrepo` under `WRK_HOME/worktrees`.
2. Run named bring under fake TTY with stdin `\n`.

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	req.StdinInput = "\n"
	return nil
}
```
