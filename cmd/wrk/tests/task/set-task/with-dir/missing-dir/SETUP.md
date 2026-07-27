# Scenario

**Feature**: wrk <non-existent-dir> --set-task errors with "does not exist"

```
wrk <nonexistent> --set-task "task" -> non-zero exit, "does not exist"
```

## Steps

1. Use a non-existent directory path as TargetDir.
2. Run `wrk <nonexistent> --set-task "task"`.
3. Verify non-zero exit with "does not exist" error.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.TargetDir = filepath.Join(req.WorkRoot, "nonexistent-dir")
	req.SetTaskDesc = "my task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```