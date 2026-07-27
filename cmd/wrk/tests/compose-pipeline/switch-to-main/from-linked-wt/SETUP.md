# Scenario

**Feature**: from a linked worktree, switch activeRoot to main via done/merge-back or `--main` pipeline partners

```
# Path A: done/merge-back switches after stage 2
linked wt -> wrk --done|--merge-back […] -> activeRoot := main for sync/tag/push/reinstall/exec

# Path B: --main + pipeline partners (no shell, no merge/remove)
linked wt -> wrk --main --sync --tag-next … -> activeRoot := main at start; wt kept
```

## Steps

- Grouping: done/merge-back success leaves at this level; `via-main-flag/` for scope rewrite without done.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
