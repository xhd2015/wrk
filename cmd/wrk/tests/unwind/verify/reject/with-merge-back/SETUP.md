# Scenario

**Feature**: `--verify` cannot be used with `--merge-back`

```
wrk --unwind --verify --merge-back
  -> Error mentions verify and merge-back
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--merge-back`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--merge-back"}
	return nil
}
```
