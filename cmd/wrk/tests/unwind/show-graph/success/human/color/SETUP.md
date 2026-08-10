# Scenario

**Feature**: human show-graph stdout color via `--color` / `--no-color` (Args only)

```
wrk --unwind --show-graph --color    -> ANSI on stdout
wrk --unwind --show-graph --no-color -> plain stdout
```

## Preconditions

- Parallel-safe: inject color mode only via `req.Args` (no `t.Setenv` / `NO_COLOR`
  in harness). Capture is non-TTY so **auto** is plain unless `--color`.
- JSON leaves stay uncolored (covered under success/json).

## Steps

1. Grouping for color force vs never.
2. Children reuse a cheap single-main dirty fixture.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```
