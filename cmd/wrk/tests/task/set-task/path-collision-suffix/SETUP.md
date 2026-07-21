# Scenario

**Feature**: --set-task path collision walks `-N` suffix (P3 T2)

```
# linked wt main-{date}-original-task; target dir for "new task" already exists as empty dir
wrk --set-task "new task" (WRK_SET_TASK_CONFIRM=1)
  -> skip occupied path/branch names until free
  -> path myrepo-main-{date}-new-task-1 + branch main-{date}-new-task-1
```

## Steps

1. Create main repo and spawn worktree with `--task "original task"`.
2. Pre-create the would-be target directory `{WRK_HOME}/worktrees/myrepo-main-{date}-new-task` (empty, not a worktree).
3. Run `wrk --set-task "new task"` with confirm env from inside the worktree.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir

	// Occupy preferred rename path so set-task must suffix-walk.
	blocked := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify("new task"), 0)
	mkdirAll(t, blocked)

	req.SetTaskDesc = "new task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
