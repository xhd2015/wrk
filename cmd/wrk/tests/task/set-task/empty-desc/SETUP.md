# Scenario

**Feature**: --set-task with empty description errors before TTY check

```
wrk --set-task "" -> non-zero exit, error about empty description
```

## Steps

1. Create a worktree.
2. Use req.Args = ["--set-task", ""] to pass the empty value.
3. Verify non-zero exit (error about empty task).

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
	req.Args = []string{"--set-task", ""}
	return nil
}
```