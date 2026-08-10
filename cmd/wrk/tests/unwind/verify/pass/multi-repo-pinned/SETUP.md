# Scenario

**Feature**: multi-repo stack with require == dep latest and no droppable replace → pass

```
# root + nested external leaf; require leaf@v0.0.1 matches tag; no replace; both clean
wrk --unwind --verify
  -> require-drift pass; droppable-replace pass; cascade-pending pass
  -> result: pass; exit 0; zero mutations
```

## Steps

1. Seed multi-repo pinned-clean fixture (no external replace).
2. Run human verify.
3. Assert full pass catalog.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifyMultiRepoPinnedClean(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
