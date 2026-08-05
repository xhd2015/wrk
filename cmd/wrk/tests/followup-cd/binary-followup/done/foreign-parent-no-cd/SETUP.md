# Scenario

**Feature**: --done auto-cd suppressed when a parent of shell cwd is a different git repo

```
# consumer (git A) holds linked worktree of myrepo (git B) at consumer/dep-wt
# shell cwd = dep-wt; --done removes it; parent consumer is foreign main
WRK_FOLLOWUP_FILE set; wrk --done
  -> wt removed; follow-up empty (foreign-repo ancestor gate)
```

## Steps

1. Init dep main `myrepo` and a separate consumer git repo.
2. Create a linked worktree of myrepo exactly at `consumer/dep-wt` (target-dir).
3. Commit on the worktree and ff-merge into myrepo (already-included path).
4. Run `wrk --done` from the worktree with follow-up env set.
5. Expect success, worktree gone, empty follow-up (no yank to myrepo main).

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
	req.CLIArgs = []string{"--done"}
	return nil
}
```
