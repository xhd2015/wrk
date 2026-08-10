# Scenario

**Feature**: `--color` colors FAIL / result fail in red on verify stdout

```
# dirty tagged main -> wrk --unwind --verify --color
  -> CSI present; dirty-peel FAIL; result: fail; exit 1
```

## Steps

1. Seed dirty tagged single main.
2. Run verify with `--color`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainDirtyTagged(t, req)
	req.Args = verifyArgs("--color")
	recordUnwindBaseline(t, req)
	return nil
}
```
