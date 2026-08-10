# Scenario

**Feature**: `--no-color` forces plain human verify stdout

```
clean main -> wrk --unwind --verify --no-color
  -> stdout has no CSI; result: pass; exit 0
```

## Steps

1. Seed clean tagged single main.
2. Run verify with `--no-color`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainClean(t, req)
	req.Args = verifyArgs("--no-color")
	recordUnwindBaseline(t, req)
	return nil
}
```
