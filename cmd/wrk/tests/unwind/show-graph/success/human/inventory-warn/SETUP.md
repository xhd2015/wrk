# Scenario

**Feature**: inventory soft warnings on show-graph (missing replace targets)

```
# missing local-replace target
show-graph -> warning: on stderr; graph still printed; exit 0
```

## Steps

1. Grouping scopes inventory-warning human leaves.

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
