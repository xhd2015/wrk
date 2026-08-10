# Scenario

**Feature**: module still owned-changed / next tag planned → owned-changed FAIL

```
# tag v0.0.1 then committed owned change; clean working tree
wrk --unwind --verify
  -> owned-changed FAIL
  -> cascade-pending may also FAIL; result: fail; exit 1
```

## Steps

1. Seed single main with owned change after tag (no DIRTY).
2. Run human verify.
3. Assert owned-changed FAIL.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainOwnedChanged(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
