# Scenario

**Feature**: `--verify` cannot be used with `--done`

```
wrk --unwind --verify --done
  -> Error mentions verify and done
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--done`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--done"}
	return nil
}
```
