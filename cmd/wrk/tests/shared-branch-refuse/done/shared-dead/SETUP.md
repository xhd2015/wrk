# Scenario

**Feature**: `wrk --done` refuses when a second (dead) registration still holds the branch

```
# wt2 directory deleted without prune; still registered on branch B
myrepo + wt1 (B) + dead wt2 (B)
  -> wrk --done from wt1
  -> non-zero Error: + paths; dead line prune hint
  -> git -C <main> worktree prune
  -> wt1 not removed
```

## Steps

1. Shared two-linked fixture, then `os.RemoveAll` second checkout (no prune).
2. Run `wrk --done` from primary wt.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedDead(t, req)
	req.Args = []string{"--done"}
	return nil
}
```
