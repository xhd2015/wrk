# Scenario

**Feature**: multi-stage compose from main without done/merge-back (activeRoot stays main)

```
# Main checkout: tag-next OK; sync/push/reinstall/gen-commit/exec compose in fixed order
myrepo (main) -> wrk --sync --tag-next --push …  (no --done)
  -> success or dry-run plan; all stages on main
```

## Steps

- Grouping only.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
