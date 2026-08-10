# Scenario

**Feature**: module status surfaces latest tag and planned next after owned change

```
# tag v0.0.1; commit owned change → next v0.0.2
wrk --unwind --show-graph
  -> module node mentions v0.0.1 (latest) and v0.0.2 (next) or owned-changed
  -> exit 0; zero mutations (no tag create)
```

## Steps

1. Seed single main with tag baseline + committed owned change + DIRTY.
2. Run show-graph.
3. ExpectedPinVersion = next tag; OldRequireVersion = latest tag.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainTaggedOwnedChange(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
