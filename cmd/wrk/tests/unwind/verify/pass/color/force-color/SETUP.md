# Scenario

**Feature**: `--color` forces ANSI on human verify pass stdout

```
clean main -> wrk --unwind --verify --color
  -> stdout contains CSI escapes (pass / banners)
  -> result: pass; exit 0
```

## Steps

1. Seed clean tagged single main.
2. Run verify with `--color`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainClean(t, req)
	req.Args = verifyArgs("--color")
	recordUnwindBaseline(t, req)
	return nil
}
```
