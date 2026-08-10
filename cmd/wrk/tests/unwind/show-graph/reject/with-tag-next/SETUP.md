# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--tag-next`

```
wrk --unwind --show-graph --tag-next
  -> Error mentions show-graph and tag-next
  -> exit ≠ 0
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with apply partner `--tag-next`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--tag-next"}
	return nil
}
```
