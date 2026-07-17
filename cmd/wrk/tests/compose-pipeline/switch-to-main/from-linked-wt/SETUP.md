# Scenario

**Feature**: `--done` / `--merge-back` from a linked worktree — activeRoot switches to main after stage 2

```
# Linked wt is legal for done/merge-back; later stages use main
linked wt -> wrk --done|--merge-back […] -> activeRoot := main for sync/tag/push/reinstall/exec
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
