# Scenario

**Feature**: piped --status without --color emits plain brief Master: labels

```
linked wt with main ahead -> wrk --status (pipe, no --color) -> no ANSI, Master: needs fast forward(...)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `main` (+1 commit ahead).
4. Run `wrk --status` (no `--color`) from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorStatusHelpersUsed()
	mainRepo := setupColorStatusMainRepo(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addColorStatusLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")
	commitColorStatusOnMain(t, mainRepo, "ahead-on-main.txt", "ahead\n", "main ahead commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```