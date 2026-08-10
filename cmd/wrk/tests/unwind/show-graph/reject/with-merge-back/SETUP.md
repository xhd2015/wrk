# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--merge-back`

```
wrk --unwind --show-graph --merge-back
  -> Error mentions show-graph and merge-back
  -> exit ≠ 0
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with land partner `--merge-back`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--merge-back"}
	return nil
}
```
