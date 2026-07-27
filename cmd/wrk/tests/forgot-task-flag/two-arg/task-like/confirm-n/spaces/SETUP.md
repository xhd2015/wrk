# Scenario

**Feature**: two-arg spaces + confirm n → fixed target-dir path create

```
WRK_TASK_LIKE_CONFIRM=1 + stdin "n\n"
  wrk <myrepo> "{WorkRoot}/out with spaces"
  -> worktree exactly at that path (parent WorkRoot exists)
  -> NOT under WRK_HOME default naming
```

## Steps

1. Init myrepo.
2. SpawnDir = absolute multi-word path under WorkRoot (parent exists, path missing).
3. Decline promote with `n`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	req.SpawnDir = filepath.Join(req.WorkRoot, "out with spaces")
	req.StdinInput = "n\n"
	return nil
}
```
