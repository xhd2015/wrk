# Scenario

**Feature**: compose without `--done` / `--merge-back` / scope-rewrite `--main` partners — `activeRoot` stays effective cwd

```
# No stage-2 switch and no --main pipeline scope rewrite:
# every requested stage runs against the starting checkout toplevel
main | linked-wt
  -> wrk [--gen-commit] [--sync] [--tag-next] [--push] [--reinstall-local] [--exec]
  -> activeRoot unchanged; --tag-next only legal when that root is main

# Exception leaf under from-main: redundant --main + partners still activeRoot=main
# (notice + continue; see main-flag-pipeline-dry-run)
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
