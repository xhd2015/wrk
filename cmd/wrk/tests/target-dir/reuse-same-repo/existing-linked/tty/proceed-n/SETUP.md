# Scenario

**Feature**: `--confirm` + `n` creates a new worktree under target when linked WT already exists

```
# existing WRK_HOME worktree remains; --confirm forces prompt; user answers n
# create proceeds as today: named subdir under existing target dir
myrepo-main-{date} exists
  -> wrk myrepo {WorkRoot}/target --confirm  (stdin n\n)
  -> new {WorkRoot}/target/myrepo-main-{date}-1
```

## Steps

1. Pre-create one linked worktree under `WRK_HOME`.
2. Run named bring under fake TTY with `--confirm` and stdin `n\n`.

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	req.Args = append(req.Args, "--confirm")
	req.StdinInput = "n\n"
	return nil
}
```
