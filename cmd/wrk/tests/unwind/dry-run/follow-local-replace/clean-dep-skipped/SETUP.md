# Scenario

**Feature**: out-of-tree followed dep is **clean** → omitted from peel; consumer dirty peels `.`

```
# sibling dep via replace; dep clean; consumer dirty
# clean followed checkout may stay in inventory for DAG but produces no peel line
clean sibling dep
  -> wrk --unwind --dry-run --tag-next --push
  -> would: peel .
  -> no would: peel for dep display
```

## Steps

1. Seed sibling replace fixture; dirtify consumer only (dep clean).
2. Dry-run with pin flags (edge may exist once follow lands even if dep clean).
3. PeelOrder = `.` only.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowCleanDepSkipped(t, req)
	// Synthetic edge may still exist after follow (consumer→clean dep) → pin flags.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push"}
	recordUnwindBaseline(t, req)
	return nil
}
```
