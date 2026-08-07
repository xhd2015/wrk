# Scenario

**Feature**: intra-repo local replace (`=> ./pkgs/shared`) never becomes a separate stack member

```
# root go.mod: replace example.com/root/shared => ./pkgs/shared
# pkgs/shared is same git toplevel as root; primary dirty
intra-repo replace
  -> wrk --unwind --dry-run
  -> would: peel .
  -> no separate peel for pkgs/shared
```

## Steps

1. Seed single main with in-tree multi-module `./pkgs/shared` local replace.
2. Dirtify primary only; dry-run without pin flags (no cross-repo edges).
3. PeelOrder = `.` only.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowIntraRepoOnly(t, req)
	// No cross-repo edges → pin flags not required.
	req.Args = []string{"--unwind", "--dry-run"}
	recordUnwindBaseline(t, req)
	return nil
}
```
