# Scenario

**Feature**: --set-task on fixed-path / non-wrk-shaped directory name errors (P3 T4)

```
# worktree spawned at fixed <target-dir> "wt" (no date in dir basename)
# branch may be wrk-shaped main-{date}, but dir name is not wrk-shaped
wrk --set-task "x" from that worktree -> non-zero; stderr explains unsupported / cannot parse
```

## Steps

1. Create main repo `myrepo` on `main`.
2. Spawn a fixed-path worktree via `wrk <myrepo> <WorkRoot>/wt` (path base is `wt`, not wrk-shaped).
3. Run `wrk --set-task "my task"` from inside that worktree with confirm env (should still fail before rename).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)

	spawnPath := filepath.Join(req.WorkRoot, "wt")
	// Two positionals from WorkRoot: source repo + fixed spawn path.
	wtDir := runWrkWithArgs(t, req, req.WorkRoot, mainRepo, spawnPath)
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "my task"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
