# Scenario

**Feature**: unique-branch smoke — `wrk --done` still succeeds when branch is not shared

```
# single wrk-managed linked wt ahead of main
myrepo + wt (unique branch)
  -> wrk --done
  -> exit 0; merge + worktree remove + branch -D
```

## Steps

1. Create single linked worktree ahead of main (no second checkout).
2. Run `wrk --done` from the worktree.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUniqueLinkedAhead(t, req)
	req.Args = []string{"--done"}
	return nil
}
```
