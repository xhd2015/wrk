# Scenario

**Feature**: `wrk --new` is the create entry (former bare create)

```
# --new selects create mode
myrepo -> wrk --new [create flags...]
  -> exit 0; stdout worktree path\n
  -> under {WRK_HOME}/worktrees/
```

## Steps

- Leaves run from a fresh main repo with `--new` present.
- Optional create modifiers (`--no-config`, `-t`, …) remain create when combined with `--new`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardMainRepo(t, req)
	return nil
}
```
