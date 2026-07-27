# Scenario

**Feature**: --fetch ignored on linked worktree cwd (no fetch, no Remote:)

```
# cwd = in-tree linked; Dir via statusDirLine (current path → ".")
linked wt cwd + --status --fetch -> no Remote; Master ok; Dir . for linked path
```


```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "linked-main")
	initFetchVerboseRepo(t, mainRepo, "linked main base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status", "--fetch"}
	return nil
}
```