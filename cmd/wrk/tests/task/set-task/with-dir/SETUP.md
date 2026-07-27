# Scenario

**Feature**: wrk <dir> --set-task renames a linked worktree at the given directory

```
# wrk resolves the target worktree from the <dir> argument instead of cwd
wrk <linked-worktree-dir> --set-task "new desc" -> parse {branchBase}-{YYYY-MM-DD}[-slug][-N]
                                                     -> git worktree move
```

## Preconditions

- `<dir>` must be a linked worktree with a wrk-naming branch.
- Process cwd is irrelevant when `<dir>` is given (the argument resolves to the effective working directory).

## Steps

- Create a worktree with --task.
- Run `wrk <wt-dir> --set-task <desc>` from WorkRoot (or any non-worktree dir).
- Verify expected outcome.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```