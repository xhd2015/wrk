# Scenario

**Feature**: `--new` is mutually exclusive with other wrk modes

```
wrk --new + --done | --list | --status
  -> non-zero; mutually exclusive; empty stdout
  -> no create worktree
```

## Steps

- Leaves combine `--new` with another mode flag from a git repo cwd.
- Mode validation must fail before create side effects.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardMainRepo(t, req)
	return nil
}
```
