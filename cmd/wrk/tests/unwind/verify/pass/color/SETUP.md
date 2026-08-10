# Scenario

**Feature**: `--color` / `--no-color` control ANSI on human verify stdout

```
clean pass stack -> wrk --unwind --verify --color|--no-color
  -> --color: CSI present; --no-color: plain
  -> still result: pass; exit 0
```

## Steps

1. Grouping scopes color leaves under pass.
2. Leaves use clean tagged main.

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
