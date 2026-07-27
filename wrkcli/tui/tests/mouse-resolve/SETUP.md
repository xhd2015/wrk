# Scenario

**Feature**: dashboard hitmap + dual-origin ResolveMouseHit product adapters

```
# layout → hitmap (view-local, real dashboard paint)
dashboard layout (width, addDisabled)
  -> BuildDashboardHitmap
  -> Hit[] + viewLines
  # stage rows: left focus | right runStage

# resolve absolute mouse → hit
absX, absY, height, viewLines, originY?
  -> ResolveMouseHit (+ hitmap, loading?)
  -> Hit | miss
  # known origin: localY = absY - originY
  # dual-origin: try top then bottom (go-pkgs mouse.Resolve)
  # prefer non-empty runStage when candidates disagree on Run column

# bug regression (top-anchored)
click gen-commit-msg Run with height >> viewLines
  -> runStage == "gen-commit-msg"
  # must NOT be "tag-next"
```

## Preconditions

- Module `github.com/xhd2015/wrk`; package under test `github.com/xhd2015/wrk/wrkcli/tui`.
- **Coverage backfill (P3 slim):** product adapters are implemented; remaining
  leaves must stay **GREEN**. Pure HitTest / dual-origin math without dashboard
  layout is owned by go-pkgs `tui/mouse/tests` — do not duplicate there.
- No git repo, PTY, iTerm2, or `tea.Program` required.
- In-memory only; default layout width 80 unless a leaf sets `req.Width`.

## Steps

1. Root sets default `req` fields for pure resolve (width, no loading).
2. Grouping nodes narrow origin mode (dual vs known), geometry (top/bottom),
   or op (`hitmap` / loading).
3. Leaves set `StageID` / `Target` (and geometry offsets); root `Run` builds
   hitmap, aims the click, calls `ResolveMouseHit` (or returns hitmap only).

## Context

- Significance order: dual-origin top-anchored bug first, then bottom-anchored,
  known-origin, hitmap layout seals, loading gate.
- `gen-commit-msg` is under Pre; `tag-next` under After — distinct local Y in
  the real dashboard layout.
- CPR chain product smoke: sibling `csi6n-origin`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Defaults for pure resolve leaves; grouping/leaf Setup narrows.
	if req.Width == 0 {
		req.Width = 80
	}
	if req.Op == "" {
		req.Op = "resolve"
	}
	if req.Target == "" && req.StageID != "" {
		req.Target = "run"
	}
	req.Loading = false
	req.AddDisabled = false
	return nil
}
```
