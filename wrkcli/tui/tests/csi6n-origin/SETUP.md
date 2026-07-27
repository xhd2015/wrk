# Scenario

**Feature**: product chain CPR → originY → dashboard ResolveMouseHit

```
# after paint: cursor on last dashboard line
TUI paint (viewLines) + injected CPR bytes
  -> tui.ParseCPR (go-pkgs re-export)
  -> tui.OriginFromCPR
  -> originY0 | fail

# product resolve with real dashboard hitmap
originY0 + BuildDashboardHitmap + abs click
  -> ResolveMouseHit(OriginY=&originY0)
  -> gen-commit-msg Run | no known origin if CPR fails
```

## Preconditions

- Module `github.com/xhd2015/wrk`; package under test `github.com/xhd2015/wrk/wrkcli/tui`.
- **Coverage backfill (P3 slim):** product adapters and layout are implemented;
  remaining leaves must stay **GREEN**. Pure ParseCPR / OriginFromCPR edges are
  **not** owned here — see go-pkgs `tui/mouse/tests`.
- No git repo, PTY, iTerm2, or `tea.Program` required.
- In-memory only; default layout width 80.

## Steps

1. Root sets no product state beyond default width.
2. Chain grouping sets `req.Op = "chain"`.
3. Leaves inject Buf / BlankAbove and stage aim; root `Run` builds hitmap,
   parses CPR, derives origin, then `ResolveMouseHit` when origin is valid.

## Context

- Sibling tree `mouse-resolve` owns hitmap layout, dual-origin stage aims, and
  loading gate.
- Pure algorithm: `external/dot-pkgs-…/go-pkgs/tui/mouse/tests` (hit-test,
  resolve, origin-from-cpr, demux, tracker).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Width == 0 {
		req.Width = 80
	}
	return nil
}
```
