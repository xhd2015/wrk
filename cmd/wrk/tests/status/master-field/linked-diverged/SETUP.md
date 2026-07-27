# Scenario

**Feature**: diverged linked worktree shows Master: diverged

```
# main and wt-side each gain unique commits
main + wt-linked -> commit on main + commit on wt-side -> wrk --status

# linked wt block compares main branch vs worktree branch
linked wt block -> Master: diverged(N commits)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `main` and on `wt-side` (one commit each, diverged).
4. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureMasterFieldHelpersUsed()
	mainRepo := setupMainRepoWithSubject(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-side")
	commitOnMain(t, mainRepo, "main-only.txt", "main\n", "main only commit")
	commitOnWorktree(t, wtDir, "wt-only.txt", "wt\n", "wt only commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```