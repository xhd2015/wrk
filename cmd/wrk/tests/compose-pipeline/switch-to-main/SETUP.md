# Scenario

**Feature**: compose switches `activeRoot` to main — via `--done`/`--merge-back` (stage 2) or via `--main` scope rewrite (start)

```
# Path A: after successful done/merge-back, stages 3–8 run with activeRoot = main
linked worktree (or rejected main)
  -> wrk [--gen-commit…] --done|--merge-back [sync tag-next push reinstall exec…]
  -> stage 2 on source; then activeRoot := main

# Path B: --main + pipeline partners (no shell, no done)
linked worktree
  -> wrk --main --sync --tag-next --push --reinstall-local [--exec] [--dry-run]
  -> activeRoot := main at start; stages on main; worktree kept
```

## Steps

- Grouping: split by cwd (linked worktree vs main). Linked-wt leaves cover done/merge-back and via-main-flag.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
