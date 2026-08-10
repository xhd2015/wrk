# Scenario

**Feature**: JSON top-level and nested keys for single dirty main

```
root (dirty) -> wrk --unwind --show-graph --json
  -> keys: repos, modules, summary, warnings
  -> repos: nodes, edges, peel_order, has_pending_edges, needs_land
  -> modules: nodes, edges
```

## Steps

1. Seed dirty single main.
2. Run show-graph with `--json`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = showGraphJSONArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
