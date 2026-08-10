# Scenario

**Feature**: missing replace target → warning: on stderr; verify exit 0 when clean

```
# root go.mod: replace => missing path; primary clean + tagged
wrk --unwind --verify
  -> stderr contains warning:
  -> error checks pass; result: pass; exit 0
```

## Steps

1. Seed consumer with local replace to a non-existent path (follow helper).
2. Remove DIRTY; tag HEAD so dirty-peel / owned-changed can pass.
3. Run verify; assert warning + pass.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowMissingTarget(t, req)
	// Follow helper dirties primary; clean for pass-path inventory-warn.
	markCleanTracked(t, req.MainRepo)
	createLightweightTag(t, req.MainRepo, unwindApplyOldTag, "HEAD")
	req.PeelOrder = nil
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
