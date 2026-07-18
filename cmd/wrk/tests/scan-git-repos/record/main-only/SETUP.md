# Scenario

**Feature**: default scan always prints main only; skips linked worktrees on stdout and in projects

```
scan-root/main (RepoTypeMain) + scan-root/main-wt (linked worktree)
  -> wrk --scan-git-repos scan-root  (no --include-worktrees)
  -> stdout is main path only (always-print main; omit worktree)
  -> records main only (source=scan)
  -> does not record worktree path
```

## Steps

1. Create main repo at `{WorkRoot}/scan-root/main`.
2. Add linked worktree `{WorkRoot}/scan-root/main-wt` (same parent root so scan can see both).
3. Run `wrk --scan-git-repos <scan-root>`.

```go
func Setup(t *testing.T, req *Request) error {
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "main")
	wtDir := setupScanLinkedWorktree(t, mainRepo, "main-wt", "scan-wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.Args = []string{"--scan-git-repos", scanRoot}
	return nil
}
```
