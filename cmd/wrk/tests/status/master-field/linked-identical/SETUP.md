# Scenario

**Feature**: linked worktree at same commit shows Master: identical

```
# linked wt created at current main HEAD with no extra commits
main + wt-linked (identical) -> wrk --status -> Master: identical
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side` (same commit as main).
3. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureMasterFieldHelpersUsed()
	mainRepo := setupMainRepoWithSubject(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-side")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```