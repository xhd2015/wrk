# Scenario

**Feature**: bare `--verify` without `--unwind` is rejected

```
wrk --verify
  -> Error names verify and unwind
  -> exit ≠ 0; no verify body
```

## Steps

1. Parent seeded clean main.
2. Run with `--verify` only (no `--unwind`).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--verify"}
	return nil
}
```
