# Scenario

**Feature**: `--no-color` forces plain human show-graph stdout

```
dirty single main -> wrk --unwind --show-graph --no-color
  -> stdout has no ANSI escapes
```

## Steps

1. Seed dirty single main.
2. Run show-graph with `--no-color`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = showGraphArgs("--no-color")
	recordUnwindBaseline(t, req)
	return nil
}
```
