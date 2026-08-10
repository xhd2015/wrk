# Scenario

**Feature**: any error-severity check FAIL → exit 1, report on stdout, no Error:

```
# residual post-ship condition (dirt / land / owned / drift / replace / cascade)
wrk --unwind --verify
  -> check id shows FAIL; result: fail; exit 1
  -> full report on stdout; no Error: for logical FAIL
  -> zero mutations
```

## Preconditions

- Each leaf focuses on a **primary** catalog id; co-failures of related checks are OK.
- Logical FAIL is not a fatal preflight: must print verify body.

## Steps

1. Grouping scopes fail leaves by primary check id (+ color FAIL).
2. Leaves assert exit 1, FAIL token, result fail, no Error: prefix.

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
