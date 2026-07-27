# Scenario

**Feature**: wrk --task when spawning a new worktree appends slug to names

```
# --task slug is appended after YYYY-MM-DD in dir and branch names
wrk --task "fix login bug" -> git worktree add -b {branch}-{date}-{slug} {path}
```

## Preconditions

- A git repo on a named branch exists in the work root.

## Steps

- Create a git repo with at least one commit.
- Run `wrk --task <desc>` from the repo directory.
- `WRK_HOME` and `WRK_DATE` are set by the test harness.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

```