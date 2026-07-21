# Scenario

**Feature**: wrk <dir> --set-task renames a worktree at the given directory

```
# create worktree from main, then rename it from another directory
wrk <linked-worktree-dir> --set-task "new task" (WRK_SET_TASK_CONFIRM=1)
  -> computes new dir/branch names
  -> git worktree move
  -> git branch -m
  -> prints new path on stdout
```

## Steps

1. Create main repo with initial commit.
2. Spawn consumer linked worktree with `--task "original task"` from main.
3. Change process cwd to WorkRoot (not the worktree).
4. Run `wrk <wt-dir> --set-task "new task"`.
5. Verify worktree moved, branch renamed, old paths gone.

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

	// Create worktree with --task
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir

	// Set repo dir to WorkRoot (not the worktree), so we test the <dir> argument.
	req.RepoDir = req.WorkRoot
	req.TargetDir = wtDir
	req.SetTaskDesc = "new task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```