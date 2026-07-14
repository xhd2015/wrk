# Scenario

**Feature**: TTY named bring with existing linked WT + `n` creates a new worktree under target

```
# existing WRK_HOME worktree remains; user answers n
# create proceeds as today: named subdir under existing target dir
myrepo-main-{date} exists
  -> wrk myrepo {WorkRoot}/target  (stdin n\n)
  -> new {WorkRoot}/target/myrepo-main-{date}
```

## Steps

1. Pre-create one linked worktree under `WRK_HOME`.
2. Run named bring under fake TTY with stdin `n\n`.

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	req.StdinInput = "n\n"
	return nil
}
```
