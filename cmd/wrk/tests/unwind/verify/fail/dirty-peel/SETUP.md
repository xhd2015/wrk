# Scenario

**Feature**: dirty stack peels → dirty-peel FAIL

```
# single main tagged at HEAD + DIRTY (no post-tag owned commits)
wrk --unwind --verify
  -> dirty-peel FAIL; result: fail; exit 1
  -> report on stdout; no Error:; zero mutations (DIRTY remains)
```

## Steps

1. Seed dirty tagged single main.
2. Run human verify.
3. Assert dirty-peel FAIL + exit 1.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainDirtyTagged(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
