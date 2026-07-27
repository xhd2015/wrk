# Scenario

**Feature**: wrk --status from linked worktree cwd omits Remote: everywhere

```
# cwd = in-tree linked wt (wt-linked); Dir labels use statusDirLine(inv cwd)
linked wt cwd -> Dir . (not main-relative wt-linked); Master where applicable; no Remote:
```


```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "from-linked")
	initFetchVerboseRepo(t, mainRepo, "from linked base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-from-linked")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status"}
	return nil
}
```