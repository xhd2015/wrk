# Scenario

**Feature**: remaining droppable external stack replace → droppable-replace FAIL

```
# root require matches leaf latest; replace => external/… still present; clean trees
wrk --unwind --verify
  -> droppable-replace FAIL
  -> cascade-pending may also FAIL (C-DR7 pin); result: fail; exit 1
```

## Steps

1. Seed external replace-only fixture with clean trees.
2. Run human verify.
3. Assert droppable-replace FAIL.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifyDroppableReplace(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
