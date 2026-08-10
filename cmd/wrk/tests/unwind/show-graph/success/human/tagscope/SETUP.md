# Scenario

**Feature**: tagscope-derived module status (latest tag / next) on show-graph

```
# MainRepo with baseline tag + owned change at HEAD
show-graph -> module status shows latest tag and next (or owned-changed)
```

## Steps

1. Grouping scopes tagscope human leaves.

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
