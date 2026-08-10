# Scenario

**Feature**: two-repo require cycle is rejected on show-graph (no success body)

```
# cycle-a requires cycle-b; cycle-b requires cycle-a; both nested dirty
A ↔ B -> wrk --unwind --show-graph
  -> Error: cycle …; exit ≠ 0
  -> no graph banners; HEADs unchanged
```

## Steps

1. Build host root linked wt with external cycle-a and cycle-b mutual requires.
2. Dirtify both cycle members (parent helper).
3. Run show-graph only (no dry-run / apply partners).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoCycleStack(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
