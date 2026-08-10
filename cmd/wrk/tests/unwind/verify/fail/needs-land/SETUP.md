# Scenario

**Feature**: linked dirty stack needing land → needs-land FAIL

```
# three-repo chain under linked consumer wt; all dirty
wrk --unwind --verify
  -> needs-land FAIL (plan.NeedsLand)
  -> dirty-peel also FAIL OK; result: fail; exit 1
```

## Steps

1. Seed three-repo linked stack; dirty all checkouts.
2. Run human verify.
3. Assert needs-land FAIL + exit 1.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifyNeedsLand(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
