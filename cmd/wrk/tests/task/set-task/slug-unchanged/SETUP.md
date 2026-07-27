# Scenario

**Feature**: --set-task with same slug is a no-op (does not trigger scan or file writes)

```
# inside linked worktree with task slug "my-task"
wrk --set-task "my-task" (WRK_SET_TASK_CONFIRM=1)
  -> detects slug unchanged
  -> prints "task unchanged"
  -> no git operations performed
```

## Steps

1. Create consumer worktree with `--task "my-task"`.
2. Run `wrk --set-task "my-task"` from inside it.
3. Verify stdout is "task unchanged", worktree path unchanged, branch unchanged.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "my-task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "my-task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
