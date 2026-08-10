# Scenario

**Feature**: sibling local-filesystem replace appears as module replace edge

```
# root replace => ../external/dot-pkgs-…; both dirty
wrk --unwind --show-graph
  -> module edge kind replace (=> path)
  -> repo nodes include dep when follow expands inventory
  -> peel free-first dep then .
```

## Steps

1. Seed sibling follow fixture (require + local replace).
2. Run show-graph (no pin flags — show-graph skips ValidateUnwindFlags).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowSiblingBothDirty(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
