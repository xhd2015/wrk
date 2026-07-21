# Scenario

**Feature**: `--exec` after successful `--set-task` runs in the renamed worktree

```
linked wt with parseable wrk name
  -> wrk --set-task <new> --exec pwd
  -> rename via git worktree move
  -> stdout: <new-path>\n<new-path>\n
  -> child cmd.Dir = new path (post-move)
```

## Preconditions

- Linked wrk worktree with date-parseable branch/dir name.
- Tests auto-confirm rename via `WRK_SET_TASK_CONFIRM=1` or `-y`.

## Steps

- Spawn worktree with `--task`, then run `--set-task` + `--exec` from inside it.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

