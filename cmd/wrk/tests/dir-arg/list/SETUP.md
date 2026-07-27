# Scenario

**Feature**: wrk --list with optional directory argument

```
# wrk <dir> --list from WorkRoot -> git worktree list for <dir>
```

## Preconditions

- Git must be available.

## Steps

- Tests invoke `wrk <repoDir> --list` via `req.Args = []string{"--list"}`.
- `req.TargetDir` is the repo passed as the first `wrk` argument.
- Expected stdout is captured with `gitWorktreeListIsolated(t, req.TargetDir)` for exact comparison.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```