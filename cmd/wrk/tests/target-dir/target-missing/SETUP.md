# Scenario

**Feature**: <target-dir> does not exist on disk — split on whether its parent exists

```
# <target-dir> absent: parent exists -> spawn exactly at <target-dir>; parent absent -> error
wrk <dir> <absent-target> -> <target-dir> (parent exists) | error (parent missing)
```

## Steps

- Leaves set `req.SpawnDir` to a path that does not exist under `{WorkRoot}`.
- `parent-exists/` is a SETUP-only grouping node (no ASSERT.md); children use `{WorkRoot}/wt`.
- `parent-missing/` uses `{WorkRoot}/missing-parent/wt`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
