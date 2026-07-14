# Scenario

**Feature**: --color wraps Master: needs fast forward in orange

```
main ahead of linked wt -> wrk --status --color -> Master: <orange>needs fast forward(+N commits)</orange>
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `main` (+1 commit ahead).
4. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
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