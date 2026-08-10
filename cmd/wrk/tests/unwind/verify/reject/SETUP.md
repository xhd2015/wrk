# Scenario

**Feature**: `--verify` is mutually exclusive with dry-run / apply partners / show-graph

```
# minimal stack + forbidden partner flag (or bare --verify)
wrk --verify | --unwind --verify --dry-run | --show-graph | …
  -> Error names verify and partner (or unwind)
  -> exit ≠ 0; no verify body; zero mutations
```

## Preconditions

- Parent verify helpers: `assertVerifyReject`, `setupVerifySingleMainClean`.
- Reject is flag-layer / early control-flow — fixture only needs a valid cwd repo.

## Steps

1. Grouping scopes mutual-exclusion leaves.
2. Each leaf pairs `--verify` with one forbidden partner (or omits `--unwind`).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Shared reject fixture: clean main is enough for flag rejection.
	setupVerifySingleMainClean(t, req)
	recordUnwindBaseline(t, req)
	return nil
}
```
