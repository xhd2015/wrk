# Scenario

**Feature**: missing replace target → warning: on stderr; graph continues

```
# root go.mod: replace => missing path; primary dirty
wrk --unwind --show-graph
  -> stderr contains warning:
  -> graph banners present; peel .
  -> exit 0; zero mutations
```

## Steps

1. Seed consumer with local replace to a non-existent path.
2. Dirtify primary; run show-graph.
3. Assert warning + graph body.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowMissingTarget(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
