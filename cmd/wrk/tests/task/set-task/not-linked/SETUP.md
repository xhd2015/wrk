# Scenario

**Feature**: --set-task from main repo (not linked worktree) errors

```
wrk --set-task "x" from main repo -> non-zero exit, error
```

## Steps

1. Create a repo (main checkout).
2. Set req.SetTaskDesc = "x".
3. Verify non-zero exit and error about linked worktree.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.SetTaskDesc = "x"
	return nil
}
```