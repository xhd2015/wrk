# Scenario

**Feature**: `wrk --merge-back` refuses when the source branch is shared

```
# same multi-checkout fixture as done
shared branch on wt1 + wt2
  -> wrk --merge-back
  -> non-zero Error: refuse --merge-back
  -> no merge; source wt kept
```

## Preconditions

- Merge-back does not remove worktree even on success; refuse must not mutate either.

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
