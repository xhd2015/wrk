# Scenario

**Feature**: `--verify` cannot be used with `--gen-commit-msg`

```
wrk --unwind --verify --gen-commit-msg
  -> Error mentions verify and gen-commit-msg
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--gen-commit-msg`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--gen-commit-msg"}
	return nil
}
```
