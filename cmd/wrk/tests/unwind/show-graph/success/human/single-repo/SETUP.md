# Scenario

**Feature**: single-repo stack shapes for human show-graph

```
main-only root -> show-graph -> one repo node + one module; peel none or .
```

## Steps

1. Grouping for single-main clean vs dirty variants.

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
