# Scenario

**Feature**: --include-worktrees extends listing to linked worktrees (list-only); flag requires --scan-git-repos

```
# with scan: print main + valid worktree; record main only
scan-root/main + scan-root/main-wt
  -> wrk --scan-git-repos --include-worktrees scan-root
  -> stdout contains main and worktree abs paths
  -> projects.json unchanged; main printed; worktree only with flag

# without scan: invalid
wrk --include-worktrees
  -> non-zero; only valid with --scan-git-repos
```

## Preconditions

- Worktree leaves reuse `setupScanLinkedWorktree` (sibling of main under scan-root).
- Cwd remains non-git `{WorkRoot}`.

## Steps

- Success leaves place main + linked worktree under scan-root and set Args with both flags.
- Error leaf sets Args to bare `--include-worktrees`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureScanGitReposHelpersUsed()
	return nil
}
```
