# Scenario

**Feature**: --done --force-cd still lands when parent is a foreign git repo

```
# consumer (git A) holds linked worktree of myrepo (git B) at consumer/dep-wt
# --force-cd bypasses foreign-repo ancestor gate
WRK_FOLLOWUP_FILE set; wrk --done --force-cd
  -> wt removed; follow-up: cd <myrepo-main>
```

## Steps

1. Same nested layout as foreign-parent-no-cd.
2. Run `wrk --done --force-cd` with follow-up env set.
3. Expect follow-up `cd <main>` (force bypass).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)

	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	target := filepath.Join(consumer, "dep-wt")
	wtDir := runWrkWithArgs(t, req, mainRepo, mainRepo, target)
	branch := branchName("main", wrkDate, 0)
	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--done", "--force-cd"}
	return nil
}
```
