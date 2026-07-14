# Scenario

**Feature**: wrk --list prints git worktree list

```
# cwd inside git checkout (main, linked worktree, or nested subpath)
wrk --list -> git worktree list stdout unchanged
```

## Preconditions

- Git must be available.

## Steps

- Tests invoke `wrk --list` via `req.Args = []string{"--list"}`.
- `req.RepoDir` is the cwd passed to git discovery (`git -C cwd worktree list`).
- Expected stdout is captured with `gitWorktreeListIsolated(t, dir)` for exact comparison.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```