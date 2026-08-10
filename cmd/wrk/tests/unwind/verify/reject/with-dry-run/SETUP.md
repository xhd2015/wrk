# Scenario

**Feature**: `--verify` cannot be used with `--dry-run`

```
wrk --unwind --verify --dry-run
  -> Error: verify … dry-run
  -> exit ≠ 0; no verify body
```

## Steps

1. Parent seeded clean main.
2. Run verify with `--dry-run`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--unwind", "--verify", "--dry-run"}
	return nil
}
```
