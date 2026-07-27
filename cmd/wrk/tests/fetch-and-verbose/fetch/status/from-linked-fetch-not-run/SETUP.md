# Scenario

**Feature**: --fetch not executed from linked worktree cwd even with -v

```
linked wt cwd + --status --fetch -v -> stderr has no fetch line
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "linked-v-main")
	initFetchVerboseRepo(t, mainRepo, "linked v main base")
	wtDir := addLinkedWorktreeInRepo(t, mainRepo, "wt-linked", "wt-branch")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--status", "--fetch", "-v"}
	return nil
}
```