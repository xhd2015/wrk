# Scenario

**Feature**: wrk <dir> --set-task with other flags is mutually exclusive

```
wrk <linked-worktree-dir> --set-task "task" --list -> non-zero exit, mutual exclusion error
```

## Steps

1. Create a worktree.
2. Run `wrk <wt-dir> --set-task "task" --list`.
3. Verify non-zero exit with mutual exclusion error.

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
	req.RepoDir = req.WorkRoot
	req.TargetDir = wtDir
	req.Args = []string{"--set-task", "my task", "--list"}
	return nil
}
```