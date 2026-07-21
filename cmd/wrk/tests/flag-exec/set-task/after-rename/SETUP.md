# Scenario

**Feature**: after `--set-task` rename, `--exec pwd` runs in the new path

```
wt with task "original" -> wrk --set-task newslug --exec pwd (WRK_SET_TASK_CONFIRM=1)
  -> new path .../myrepo-main-2026-06-30-newslug
  -> stdout: newpath\nnewpath\n
```

## Steps

1. Create main repo and linked worktree with `--task original`.
2. Run `wrk --set-task newslug --exec pwd` from inside the worktree with confirm env.

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

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "newslug"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.Args = []string{"--exec", "pwd"}
	return nil
}
```
