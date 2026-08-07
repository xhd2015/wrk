# Scenario

**Feature**: replace into nested module subdir of multi-module dep → peel **dep git toplevel**

```
# dep main at ../external/dep-main-DATE with nested/ go.mod (example.com/dep/nested)
# consumer replace => ../external/dep-main-DATE/nested  (points at nested module dir)
# both dirty
replace nested module subdir
  -> wrk --unwind --dry-run --tag-next --push
  -> would: peel ../external/dep-main-2026-06-30   # ShowToplevel of nested path
  -> would: peel .
  -> NOT would: peel …/nested alone as stack Path
```

## Steps

1. Seed multi-module dep under external; consumer replace targets nested subdir.
2. Dirtify dep toplevel + consumer; dry-run with pin flags.
3. PeelOrder uses dep **git toplevel** display, not nested-only path.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowNestedModTargetToplevel(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push"}
	recordUnwindBaseline(t, req)
	return nil
}
```
