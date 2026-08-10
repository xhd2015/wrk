# Scenario

**Feature**: PlanUnwindCascade would still emit tag-next/pin → cascade-pending FAIL

```
# root + pkgs/shared; shared owned-changed after tag; clean working tree
wrk --unwind --verify
  -> cascade-pending FAIL
  -> owned-changed also FAIL OK; result: fail; exit 1
```

## Steps

1. Seed shared owned-changed single-repo fixture (no DIRTY).
2. Run human verify.
3. Assert cascade-pending FAIL.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySharedOwnedChanged(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
