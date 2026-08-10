# Scenario

**Feature**: `--json` show-graph emits stable snake_case schema (no ANSI)

```
wrk --unwind --show-graph --json
  -> JSON object: repos, modules, summary, warnings
  -> repos.peel_order array; no human banners
```

## Preconditions

- JSON allowed only with show-graph for unwind (G6).
- Leaves use showGraphJSONArgs().

## Steps

1. Grouping scopes JSON success leaves.

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
