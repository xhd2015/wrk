# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--gen-commit-msg`

```
wrk --unwind --show-graph --gen-commit-msg
  -> Error mentions show-graph and gen-commit-msg
  -> exit ≠ 0
```

## Steps

1. Parent seeded single dirty main.
2. Run show-graph with commit partner `--gen-commit-msg`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--show-graph", "--gen-commit-msg"}
	return nil
}
```
