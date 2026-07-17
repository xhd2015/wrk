# Scenario

**Feature**: compose without `--done` / `--merge-back` — `activeRoot` stays effective cwd for all stages

```
# No stage-2 switch: every requested stage runs against the starting checkout toplevel
main | linked-wt
  -> wrk [--gen-commit] [--sync] [--tag-next] [--push] [--reinstall-local] [--exec]
  -> activeRoot unchanged; --tag-next only legal when that root is main
```

## Steps

- Grouping: split by cwd (main vs linked worktree).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
