# Scenario

**Feature**: CPR → originY → ResolveMouseHit (P1 known-origin path)

```
# injected CPR as after paint on last line
buf (or synthesized from BlankAbove + viewLines)
  -> ParseCPR
  -> OriginFromCPR(viewLines)
  -> ResolveMouseHit(OriginY=&originY0)
  -> hit | no-origin fallback signal
```

## Preconditions

- Leaves set `req.Op = "chain"`.
- Uses product `BuildDashboardHitmap` / `ResolveMouseHit` plus CPR adapters
  (`ParseCPR` / `OriginFromCPR` re-exports). Pure CPR edges: go-pkgs mouse tests.
- Default stage aim is gen-commit-msg Run unless a leaf overrides.

## Steps

1. Set operation to chain.
2. Leaves either inject Buf or set BlankAbove for synthetic CPR.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "chain"
	if req.StageID == "" {
		req.StageID = "gen-commit-msg"
	}
	if req.Target == "" {
		req.Target = "run"
	}
	return nil
}
```
