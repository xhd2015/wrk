# Scenario

**Feature**: cycle reject has no successful cascade body (C-DR5)

```
# mutual require cycle among stack repos
  -> wrk --unwind --dry-run --tag-next --push --done
  -> Error mentioning cycle; exit ≠ 0
  -> no multi-step peel plan; no cascade would: tag-next body
  -> zero mutations
```

## Preconditions

- Reuses `setupTwoCycleStack` from root helpers.
- Status-quo cycle reject may stay **GREEN**; cascade absence must hold after P1.

## Steps

1. Grouping scopes cycle rejects under cascade dry-run.

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
