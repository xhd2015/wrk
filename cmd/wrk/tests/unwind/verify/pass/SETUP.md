# Scenario

**Feature**: all error-severity checks pass → exit 0, report body, zero mutations

```
# clean fully-shipped stack (tags at HEAD; no dirt; no drift; no droppable replace)
wrk --unwind --verify
  -> all 6 checks pass; result: pass; exit 0
  -> optional soft warning: on stderr still exit 0
```

## Preconditions

- Fixtures leave working trees clean and tagscope latest aligned with HEAD.
- No residual cascade (owned-changed / require-drift / droppable external replace).

## Steps

1. Grouping scopes pass leaves (clean-stack, multi-repo, inventory-warn, color).
2. Leaves assert full pass catalog + banners + zero mutations.

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
