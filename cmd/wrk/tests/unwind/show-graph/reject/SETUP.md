# Scenario

**Feature**: `--show-graph` is mutually exclusive with `--dry-run` and apply partners

```
# minimal stack + forbidden partner flag
wrk --unwind --show-graph --dry-run | --tag-next | --done | …
  -> Error names show-graph and partner flag
  -> exit ≠ 0; no graph body; zero mutations
```

## Preconditions

- Parent show-graph helpers: `assertShowGraphReject`, `setupSingleMainDirty`.
- Reject is flag-layer / early control-flow — fixture only needs a valid cwd repo.

## Steps

1. Grouping scopes mutual-exclusion leaves.
2. Each leaf pairs `--show-graph` with one forbidden partner.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Shared reject fixture: single dirty main is enough for flag rejection.
	setupSingleMainDirty(t, req)
	recordUnwindBaseline(t, req)
	return nil
}
```
