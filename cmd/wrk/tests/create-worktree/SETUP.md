# Scenario

**Feature**: wrk succeeds when cwd is a git checkout

```
# cwd inside git checkout (main or linked worktree)
wrk -> git worktree add -> stdout prints target path
```

## Preconditions

- Git must be available.

## Steps

- All tests in this branch run `wrk` with no arguments from a git checkout.
- `WRK_HOME` is isolated per test at `{WorkRoot}/.wrk`.
- Stdout must contain only the absolute worktree path (one line).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```