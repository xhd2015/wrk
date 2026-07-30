# Scenario

**Feature**: `wrk --done` refuses when the worktree branch is shared across checkouts

```
# shared live / dead / dry-run → Error: refuse --done
# unique branch smoke → --done still succeeds
linked wt on branch shared with second checkout
  -> wrk --done [--dry-run]
  -> non-zero Error: (or unique: merge-back --rm success)
```

## Preconditions

- Uses root `setupSharedTwoLinked` / `setupSharedDead` / `setupUniqueLinkedAhead`.
- Default auto-yes for ahead merge does not apply when refuse fires first.

## Steps

- Grouping only; leaves set Args and fixture.

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
