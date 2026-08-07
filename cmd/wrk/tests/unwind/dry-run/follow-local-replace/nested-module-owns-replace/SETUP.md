# Scenario

**Feature**: nested consumer module owns the out-of-tree replace (root go.mod has none)

```
# root go.mod: no replace
# root/svc/go.mod: require dep + replace => ../../external/dot-pkgs-main-DATE
# dep dirty + primary dirty
nested module owns replace
  -> wrk --unwind --dry-run --tag-next --push
  -> would: peel ../external/dot-pkgs-main-2026-06-30
  -> would: peel .
```

## Steps

1. Seed sibling dep + consumer with nested `svc/` module owning the local replace.
2. Dirtify dep + primary; dry-run with pin flags.
3. PeelOrder = dep display then `.` (follow must scan nested go.mod).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowNestedModuleOwnsReplace(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push"}
	recordUnwindBaseline(t, req)
	return nil
}
```
