# Scenario

**Feature**: clean single main — graph lists nodes; peel order empty/(none)

```
# main-only root; clean; already on main
root (clean) -> wrk --unwind --show-graph
  -> repo + module nodes present
  -> peel order (none)
  -> exit 0 without pin/land flags; zero mutations
```

## Steps

1. Seed clean single main (no DIRTY).
2. Run show-graph only.
3. PeelOrder empty.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainClean(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
