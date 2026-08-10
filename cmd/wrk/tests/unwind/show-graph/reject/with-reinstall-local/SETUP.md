# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--reinstall-local`

```
wrk --unwind --show-graph --reinstall-local
  -> Error mentions show-graph and reinstall-local
  -> exit ≠ 0
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with ship tail partner `--reinstall-local`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--reinstall-local"}
	return nil
}
```
