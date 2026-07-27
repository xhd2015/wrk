# Scenario

**Feature**: wrk succeeds when cwd is a git checkout

```
# cwd inside git checkout (main or linked worktree)
wrk --new -> git worktree add -> stdout prints target path
```

## Preconditions

- Git must be available.

## Steps

- All tests in this branch run `wrk --new` from a git checkout (P1: create entry is `--new`).
- `WRK_HOME` is isolated per test at `{WorkRoot}/.wrk`.
- Stdout must contain only the absolute worktree path (one line).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	// P1: bare no-args is dashboard; create uses --new.
	req.Args = []string{"--new"}
	return nil
}
```
