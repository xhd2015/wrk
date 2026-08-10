# Scenario

**Feature**: `--verify` cannot be used with `--reinstall-local`

```
wrk --unwind --verify --reinstall-local
  -> Error mentions verify and reinstall-local
  -> exit ≠ 0
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--reinstall-local`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--reinstall-local"}
	return nil
}
```
