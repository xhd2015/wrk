# Scenario

**Feature**: --color wraps Master: needs merge back in orange

```
linked wt ahead of main -> wrk --status --color -> Master: <orange>needs merge back(+N commits)</orange>
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `wt-side` (+1 commit ahead).
4. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
	mainRepo := setupColorStatusMainRepo(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addColorStatusLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")
	commitColorStatusOnWorktree(t, wtDir, "ahead-on-wt.txt", "ahead\n", "wt ahead commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```