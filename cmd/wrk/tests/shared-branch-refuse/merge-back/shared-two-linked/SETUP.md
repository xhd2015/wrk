# Scenario

**Feature**: `wrk --merge-back` refuses shared source branch (two live linked checkouts)

```
myrepo + wt1 (B) + wt2 (--force B)
  -> wrk --merge-back from wt1
  -> non-zero; Error: refuse --merge-back
  -> no merge; both wts remain
```

## Steps

1. Shared two-linked fixture (ahead commit on primary).
2. Run `wrk --merge-back` from primary wt.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedTwoLinked(t, req)
	req.Args = []string{"--merge-back"}
	return nil
}
```
