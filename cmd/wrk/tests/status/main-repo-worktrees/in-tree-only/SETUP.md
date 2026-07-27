# Scenario

**Feature**: in-tree linked worktree is scan-only; no append section

```
# git worktree add under main repo tree (discovered by scan)
myrepo + wt-linked (in-tree) -> wrk --status -> scan blocks only, no append
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Add in-tree linked worktree at `myrepo/wt-linked` on branch `wt-side`.
3. Run `wrk --status` from the main repo root.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initMainRepo(t, mainRepo, "in-tree only main")
	wtDir := addInTreeLinkedWorktree(t, mainRepo, "wt-linked", "wt-side")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	req.InTreeWtDir = wtDir
	req.InTreeWtRel = "wt-linked"
	req.InTreeWtBranch = "wt-side"
	return nil
}
```