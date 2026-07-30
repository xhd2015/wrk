# Scenario

**Feature**: wrk commit path refuses when HEAD branch is shared across worktrees

```
# staged changes on shared branch
shared B on wt1 + wt2; stage change.go on wt1
  -> wrk --gen-commit-msg --commit --dry-run
  -> non-zero Error: refuse commit (or gen-commit-msg / --commit)
  -> no new commit; HEAD subject unchanged
```

## Preconditions

- Uses `--gen-commit-msg --commit --dry-run` so the guard can fire without a live LLM
  (fail-closed: dry-run is still hard error when branch is shared).
- Staged file only; no agent mock required if refuse is early.

## Steps

- Grouping only.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureSharedBranchRefuseHelpersUsed()
	return nil
}
```
