# Scenario

**Feature**: already-on-main dirty repo peels without land (tag-next + push only)

```
# sole root main: v0.0.1 + owned change + DIRTY + bare origin; no stack edges
root (main, dirty)
  -> wrk --unwind --tag-next --push
  -> skip land (already main)
  -> tag v0.0.2 @ HEAD; push main+tag to bare origin
  -> exit 0; no --done/--merge-back required
```

## Steps

1. Seed single main with root-bump tag lineage + bare origin + DIRTY.
2. Run `--unwind --tag-next --push` (no land flags).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyAlreadyMainRootBump(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
