# Scenario

**Feature**: `--done` / `--merge-back` from main checkout is hard-gated (linked worktree required)

```
# main checkout is never legal for done/merge-back (even with other stages or dry-run)
myrepo (main) -> wrk --done|--merge-back […]
  -> non-zero
  -> stderr Error/wrk: names the flag and linked worktree requirement
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
