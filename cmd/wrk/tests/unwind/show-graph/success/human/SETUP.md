# Scenario

**Feature**: polished human-format show-graph (dir keys, collapsed →, replaced)

```
wrk --unwind --show-graph
  -> ==== unwind graph (repo) ====
  -> ==== unwind graph (module) ====  # dir / modules @ / → / replaced
  -> ==== status summary ====
```

## Preconditions

- No `--json`. Children set stack fixtures and PeelOrder.
- Human polish contract: module identity = dir (multi-repo label/dir); edges
  collapsed with unicode `→`; word `replaced` only; optional `(latest`;
  color leaves under `color/` use Args flags only.

## Steps

1. Grouping marks human output family.

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
