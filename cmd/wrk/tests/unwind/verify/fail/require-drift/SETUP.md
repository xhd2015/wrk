# Scenario

**Feature**: stack require version ≠ dep latest tag → require-drift FAIL

```
# sibling follow: require v0.0.0; dep tagged v0.0.1; working trees cleaned
wrk --unwind --verify
  -> require-drift FAIL
  -> cascade-pending may also FAIL; result: fail; exit 1
```

## Steps

1. Seed sibling local-replace stack; tag dep at v0.0.1; clean DIRTY.
2. Run human verify.
3. Assert require-drift FAIL.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifyRequireDrift(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
