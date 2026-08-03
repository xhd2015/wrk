# Scenario

**Feature**: fully staged dirt with gen-commit (no `--add-all`) omits leave-N

```
# sole main; DIRTY staged (index has changes; no unstaged/untracked leftover)
root (fully staged dirt) -> wrk --unwind --dry-run --gen-commit-msg --commit
  -> would: peel .
  ->   would: generate commit message and commit staged changes
  -> no leave-N line
  -> exit 0; zero mutations (index/worktree unchanged by dry-run)
```

## Steps

1. Seed single dirty main (untracked `DIRTY`).
2. `git add -A` so porcelain is fully staged (N=0 not-fully-staged).
3. Run unwind dry-run with gen-commit + commit; omit `--add-all`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	stageAll(t, req.RepoDir)
	req.LeaveN = 0
	req.Args = []string{
		"--unwind", "--dry-run",
		"--gen-commit-msg", "--commit",
	}
	recordUnwindBaseline(t, req)
	return nil
}
```
