# Scenario

**Feature**: `--set-task` with long description under long basename fits names ≤255 bytes

```
# existing linked wt under long basename (short initial task)
wrk --set-task <long desc> (WRK_SET_TASK_CONFIRM=1)
  -> rename path/branch with fitted slug; Base/branch ≤255
  -> Base == basename + "-" + branch
```

## Steps

1. Create long-basename (180) main repo.
2. Spawn worktree with short `--task "t1"` (fits without extra budget trim).
3. Run `--set-task` with long description that would overflow with soft-cap 64 slug.
4. Auto-confirm via `WRK_SET_TASK_CONFIRM=1`.

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	basename := strings.Repeat("r", 180)
	mainRepo := filepath.Join(req.WorkRoot, basename)
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/longrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	// Short initial task so create succeeds without depending on new budget fit.
	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "t1")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "explore the integration of distributed tracing with opentelemetry across all microservices and platforms"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
