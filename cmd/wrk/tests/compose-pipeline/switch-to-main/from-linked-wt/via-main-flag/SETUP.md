# Scenario

**Feature**: `--main` + pipeline partners from a linked worktree — scope rewrite to main (no shell, no done/merge/remove)

```
# Linked wt ahead; partners present → not nested shell
linked wt
  -> wrk --main --sync --tag-next --push --reinstall-local [--exec] [--dry-run]
  -> activeRoot := main at start
  -> stages run on main in fixed order
  -> worktree remains; no merge --ff-only / worktree remove
```

## Steps

- Grouping only; leaves set fixtures and Args.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
