# Scenario

**Feature**: --main --status from in-tree linked wt uses full main status (not linked shortcut)

```
# in-tree linked under main (git worktree add)
myrepo + wt-linked -> cwd = wt-linked
wrk --main --status -> full main scan (`.` + Remote + relative linked block)
  != plain wrk --status from wt-linked (single `.` + Master only)
  == wrk --status from main
```

## Steps

1. Initialize main repo; add in-tree linked worktree at `wt-linked` on branch `wt-side`.
2. cwd = in-tree linked root; Args = `--main`, `--status`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, mainRepo, "main flag in-tree linked")
	wtDir := addInTreeLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")

	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = "wt-side"
	req.RepoDir = wtDir
	setMainStatusArgs(req, "--main", "--status")
	return nil
}
```