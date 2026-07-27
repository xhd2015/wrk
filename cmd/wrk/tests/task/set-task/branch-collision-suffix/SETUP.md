# Scenario

**Feature**: --set-task branch collision walks `-N` suffix (P3 T3)

```
# linked wt main-{date}-original-task; preferred branch main-{date}-new-task already exists
wrk --set-task "new task" (WRK_SET_TASK_CONFIRM=1)
  -> path+branch get -1 until free
  -> myrepo-main-{date}-new-task-1 / main-{date}-new-task-1
```

## Steps

1. Create main repo and spawn worktree with `--task "original task"`.
2. Pre-create branch `main-{date}-new-task` in the main repo (ref only).
3. Run `wrk --set-task "new task"` with confirm env from inside the worktree.

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
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir

	// Occupy preferred rename branch so set-task must suffix-walk.
	runGitIsolated(t, mainRepo, "branch", branchNameWithTask("main", wrkDate, slugify("new task"), 0))

	req.SetTaskDesc = "new task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
