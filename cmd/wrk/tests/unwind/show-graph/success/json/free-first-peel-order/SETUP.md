# Scenario

**Feature**: JSON peel_order is free-first display-path array for multi dirty stack

```
# 3-repo all dirty
wrk --unwind --show-graph --json
  -> repos.peel_order = [leaf external, mid external, "."]
```

## Steps

1. Build three-repo dirty chain.
2. Run show-graph JSON; PeelOrder display paths free-first.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	setPeelOrderDisplays(t, req, req.DepsLinkedWtDir, req.ExternalWtDir, req.WtDir)
	req.Args = showGraphJSONArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
