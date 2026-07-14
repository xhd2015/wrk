# Scenario

**Feature**: linked worktree ahead of main shows Master: needs merge back

```
# worktree gains a commit after creation; main stays behind
main + wt-linked -> commit on wt-side -> wrk --status

# linked wt block compares main branch vs worktree branch
linked wt block -> Master: needs merge back(+N commits)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `wt-side` (+1 commit ahead of main).
4. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureMasterFieldHelpersUsed()
	mainRepo := setupMainRepoWithSubject(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-side")
	commitOnWorktree(t, wtDir, "ahead-on-wt.txt", "ahead\n", "wt ahead commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```