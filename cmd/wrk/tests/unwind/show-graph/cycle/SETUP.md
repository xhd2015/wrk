# Scenario

**Feature**: show-graph rejects stack-repo DAG cycles before printing a graph body

```
# mutual require A ↔ B
wrk --unwind --show-graph
  -> exit ≠ 0; message mentions cycle
  -> no successful graph banners; zero mutations
```

## Preconditions

- Parent helpers: `setupTwoCycleStack`, `assertCycleError`, `assertNoSuccessfulShowGraphBody`.

## Steps

1. Grouping scopes show-graph cycle preflight (no dry-run / apply flags).

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
