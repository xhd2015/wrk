# Scenario

**Feature**: cycle preflight rejects verify before report body

```
# two-repo require cycle in stack
A ↔ B -> wrk --unwind --verify
  -> Error: cycle …; exit ≠ 0
  -> no verify banners; HEADs unchanged
```

## Preconditions

- Reuses root `setupTwoCycleStack`.
- Cycle is fatal preflight (same class as show-graph / dry-run).

## Steps

1. Grouping scopes cycle leaves under verify.
2. Leaf builds mutual-require external pair and runs verify.

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
