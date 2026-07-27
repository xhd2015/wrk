# Scenario

**Feature**: --set-task migrates legacy slash branch to sanitized form (P3 T5 / P1 S2)

```
# construct legacy layout via raw git:
#   dir  {WRK_HOME}/worktrees/myrepo-feature-foo-{date}
#   branch feature/foo-{date}  (slash preserved — pre-policy)
wrk --set-task "bar" (WRK_SET_TASK_CONFIRM=1)
  -> branch becomes feature-foo-{date}-bar (no slash)
  -> dir  myrepo-feature-foo-{date}-bar
```

## Steps

1. Create main repo.
2. Manually add a linked worktree with legacy slash branch and wrk-shaped sanitized path.
3. Run `wrk --set-task "bar"` from that worktree.

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

	// Legacy branch keeps slash; path token was already sanitized historically.
	legacyBranch := "feature/foo-" + wrkDate
	token := sanitizeBranchToken("feature/foo")
	wtDir := worktreePath(req.WrkHome, "myrepo", token, wrkDate, 0)
	mkdirAll(t, filepath.Dir(wtDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", legacyBranch, wtDir)

	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "bar"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```
