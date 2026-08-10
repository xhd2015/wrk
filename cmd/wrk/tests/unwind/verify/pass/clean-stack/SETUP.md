# Scenario

**Feature**: clean single main tagged at HEAD → all verify checks pass

```
# sole main, clean, tag v0.0.1 at HEAD
wrk --unwind --verify
  -> dirty-peel..cascade-pending all pass
  -> result: pass; exit 0; trailing \n; zero mutations
```

## Steps

1. Seed clean tagged single main.
2. Run human verify.
3. Assert banners, all pass, result pass.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainClean(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
