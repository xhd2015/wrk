# Scenario

**Feature**: compose includes `--done` or `--merge-back` so stage 2 switches `activeRoot` to main

```
# After successful done/merge-back, stages 3–8 run with activeRoot = main
linked worktree (or rejected main)
  -> wrk [--gen-commit…] --done|--merge-back [sync tag-next push reinstall exec…]
  -> stage 2 on source; then activeRoot := main
```

## Steps

- Grouping: split by cwd (linked worktree vs main).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: git required for descendant compose fixtures.
	skipIfNoGit(t)
	return nil
}
```
