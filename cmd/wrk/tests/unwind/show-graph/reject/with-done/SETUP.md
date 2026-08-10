# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--done`

```
wrk --unwind --show-graph --done
  -> Error mentions show-graph and done
  -> exit ≠ 0
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with land partner `--done`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--done"}
	return nil
}
```
