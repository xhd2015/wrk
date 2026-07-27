# Scenario

**Feature**: --color wraps clean Status value in green on linked worktree

```
clean linked wt -> wrk --status --color -> Status: <green>clean</green>
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureColorStatusHelpersUsed()
	withStatusColor(req)
	mainRepo := setupColorStatusMainRepo(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addColorStatusLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```