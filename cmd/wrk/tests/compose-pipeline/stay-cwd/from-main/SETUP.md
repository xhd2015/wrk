# Scenario

**Feature**: multi-stage compose from main without done/merge-back (activeRoot stays main)

```
# Main checkout: tag-next OK; sync/push/reinstall/gen-commit/exec compose in fixed order
myrepo (main) -> wrk --sync --tag-next --push …  (no --done)
  -> success or dry-run plan; all stages on main

# Optional redundant --main with pipeline partners: notice + continue (still stay-cwd main)
myrepo (main) -> wrk --main --tag-next … --dry-run
  -> stderr notice --main not necessary; pipeline still runs
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
