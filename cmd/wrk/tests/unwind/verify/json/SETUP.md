# Scenario

**Feature**: `wrk --unwind --verify --json` emits snake_case report JSON

```
# pass or fail stack
wrk --unwind --verify --json
  -> pure JSON: work_dir, checks[6], summary, warnings
  -> no ANSI; no human banners
  -> pass → exit 0 result pass; fail → exit 1 result fail
```

## Preconditions

- JSON never includes ANSI.
- Check ids match catalog; summary.checks == 6.

## Steps

1. Grouping scopes JSON leaves (shape-keys pass + fail-status).
2. Leaves set `verifyJSONArgs()`.

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
