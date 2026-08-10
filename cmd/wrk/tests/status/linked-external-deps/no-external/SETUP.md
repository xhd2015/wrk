# Scenario

**Feature**: linked consumer with no `./external` deps — single status block (regression)

```
# linked consumer only (no external/*)
consumerWt -> wrk --status -> one Dir: . block (existing from-linked-cwd semantics)
```

## Steps

- Descendants create linked consumer only and run status from that cwd.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```
