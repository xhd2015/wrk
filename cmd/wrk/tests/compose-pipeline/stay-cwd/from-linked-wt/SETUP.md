# Scenario

**Feature**: multi-stage compose from linked worktree without done/merge-back (activeRoot stays WT)

```
# Linked wt: tag-next forbidden; other stages OK on the worktree branch/modules
linked wt -> wrk --push --sync --reinstall-local …  (no --tag-next, no --done)
  -> stages use WT activeRoot
linked wt -> wrk … --tag-next … (no --done/--merge-back)
  -> non-zero; --tag-next requires main repository
```

## Steps

- Grouping only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
