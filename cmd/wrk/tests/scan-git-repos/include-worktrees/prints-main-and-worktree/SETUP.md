# Scenario

**Feature**: --scan-git-repos --include-worktrees prints main and linked worktree; records main only

```
scan-root/main (RepoTypeMain) + scan-root/main-wt (linked worktree)
  -> wrk --scan-git-repos --include-worktrees scan-root
  -> stdout contains main abs path
  -> stdout contains worktree abs path
  -> projects.json unchanged; print main (+ worktree with flag)
```

## Steps

1. Create main repo at `{WorkRoot}/scan-root/main`.
2. Add linked worktree `{WorkRoot}/scan-root/main-wt`.
3. Run `wrk --scan-git-repos --include-worktrees <scan-root>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	scanRoot := makeScanRoot(t, req.WorkRoot)
	mainRepo := initScanMainRepo(t, scanRoot, "main")
	wtDir := setupScanLinkedWorktree(t, mainRepo, "main-wt", "scan-include-wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.Args = []string{"--scan-git-repos", "--include-worktrees", scanRoot}
	return nil
}
```
