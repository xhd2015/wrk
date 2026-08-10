# Scenario

**Feature**: soft inventory warnings do not fail verify when checks pass

```
# missing replace target → warning: on stderr
wrk --unwind --verify
  -> warning: on stderr
  -> checks still pass (or only soft) → exit 0 when no error-severity FAIL
```

## Steps

1. Grouping scopes soft-warn pass leaves.
2. Leaf reuses follow missing-target fixture and cleans dirt.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```
