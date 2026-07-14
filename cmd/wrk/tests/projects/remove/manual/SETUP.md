# Scenario

**Feature**: wrk --rm resolves directory to main repo

```
wrk --rm <dir> -> projects.json entry deleted + stdout resolved main repo path (when removed)
```

## Steps

- Descendants vary whether `<dir>` is main repo or linked worktree.

```go
func Setup(t *testing.T, req *Request) error {
	ensureRemoveHelpersUsed()
	return nil
}
```
