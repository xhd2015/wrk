# Scenario

**Feature**: `--color` forces ANSI on human show-graph stdout

```
dirty single main -> wrk --unwind --show-graph --color
  -> stdout contains CSI escapes (banners/dirty tokens)
```

## Steps

1. Seed dirty single main.
2. Run show-graph with `--color`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = showGraphArgs("--color")
	recordUnwindBaseline(t, req)
	return nil
}
```
