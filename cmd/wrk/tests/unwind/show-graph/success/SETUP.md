# Scenario

**Feature**: acyclic stack → show-graph prints repo + module graph (exit 0, zero mutations)

```
# acyclic inventory
wrk --unwind --show-graph [--json]
  -> human banners OR stable JSON
  -> peel = dirty free-first only; clean nodes still listed
  -> no pin/land flags required; zero mutations
```

## Preconditions

- Parent helpers for fixtures and graph asserts.
- Success path never passes apply partners or `--dry-run`.

## Steps

1. Grouping scopes success outcomes; children split human vs json and stack shapes.

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
