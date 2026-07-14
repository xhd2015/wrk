# Scenario

**Feature**: `wrk --set-task "new" -y` renames on non-TTY without terminal error

```
linked wt with task slug -> wrk --set-task "new task" -y (non-TTY) -> exit 0; renamed
```

## Steps

1. Create worktree with `--task "original task"`.
2. Run `wrk --set-task "new task" -y` from inside the worktree (non-TTY).

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
	req.SetTaskDesc = "new task"
	req.Args = []string{"-y"}
	return nil
}
```
