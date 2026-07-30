# Scenario

**Feature**: `wrk --done` refuses when two live linked worktrees share the branch

```
# wt1 + wt2 same branch via worktree add --force
myrepo + wt1 (branch B) + wt2 (--force B)
  -> wrk --done from wt1
  -> non-zero; Error: branch B multi-checkout; refuse --done
  -> neither wt removed; branch kept; main not merged
```

## Steps

1. Create wrk-managed linked worktree; force-add second checkout on same branch.
2. Commit ahead on primary wt.
3. Run `wrk --done` from primary wt.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedTwoLinked(t, req)
	req.Args = []string{"--done"}
	return nil
}
```
