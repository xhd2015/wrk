# Scenario

**Feature**: when `-t` is already set, second positional stays target-dir (no treat-as-task)

```
wrk <myrepo> "{WorkRoot}/wt with spaces" -t "other task"
  -> fixed path create at target; branch slug from -t; no promote of second arg
```

## Steps

1. Init myrepo.
2. SpawnDir multi-word path (would be task-like alone).
3. TaskDesc set so Run injects `--task` / `-t`.

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
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt with spaces")
	req.TaskDesc = "other task"
	req.TaskFlag = "-t"
	return nil
}
```
