# Scenario

**Feature**: linked worktree behind main shows Master: needs fast forward

```
# main gains a commit after linked wt was created
main + wt-linked (same base) -> commit on main -> wrk --status

# linked wt block compares main branch vs worktree branch
linked wt block -> Master: needs fast forward(+N commits)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Commit on `main` (+1 commit ahead of the worktree branch).
4. Run `wrk --status` from the main repo root.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureMasterFieldHelpersUsed()
	mainRepo := setupMainRepoWithSubject(t, req.WorkRoot, "myrepo", "status main root")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-side")
	commitOnMain(t, mainRepo, "ahead-on-main.txt", "ahead\n", "main ahead commit")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	return nil
}
```