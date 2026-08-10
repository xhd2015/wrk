# Scenario

**Feature**: `--verify` cannot be used with `--push`

```
wrk --unwind --verify --push
  -> Error mentions verify and push
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--push`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--push"}
	return nil
}
```
