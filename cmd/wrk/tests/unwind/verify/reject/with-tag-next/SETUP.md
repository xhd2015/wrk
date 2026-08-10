# Scenario

**Feature**: `--verify` cannot be used with `--tag-next`

```
wrk --unwind --verify --tag-next
  -> Error mentions verify and tag-next
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with apply partner `--tag-next`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--tag-next"}
	return nil
}
```
