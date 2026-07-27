# Scenario

**Feature**: --color wraps Master: diverged in red

```
main and linked wt diverged -> wrk --status --color -> Master: <red>diverged(N commits)</red>
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `main` and on `wt-side` (one commit each).
4. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
	mainRepo := setupColorStatusMainRepo(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addColorStatusLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")
	commitColorStatusOnMain(t, mainRepo, "main-only.txt", "main\n", "main only commit")
	commitColorStatusOnWorktree(t, wtDir, "wt-only.txt", "wt\n", "wt only commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```