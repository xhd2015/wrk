# Scenario

**Feature**: `--verify` cannot be used with `--show-graph`

```
wrk --unwind --verify --show-graph
  -> Error: verify … show-graph
  -> exit ≠ 0; no verify body
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--show-graph`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--show-graph"}
	return nil
}
```
