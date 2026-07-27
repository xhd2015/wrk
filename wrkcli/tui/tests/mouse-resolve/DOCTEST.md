# Mouse resolve — dashboard hitmap + dual-origin product adapters (`wrkcli/tui`)

## Version
0.0.2

**Plan phase P3 (slim):** product doctests for dashboard mouse hit resolution
in `github.com/xhd2015/wrk/wrkcli/tui`.

This tree owns **layout-specific** hitmap (`BuildDashboardHitmap`) and
**ResolveMouseHit** stages (known origin, dual-origin with real stage geometry,
loading gate). Pure dual-origin / HitTest algorithm without dashboard layout is
sealed in go-pkgs:

| Concern | Owner |
|---------|--------|
| HitTest, Resolve dual/known on synthetic hits, OriginFromCPR | `external/dot-pkgs-…/go-pkgs/tui/mouse/tests` |
| Headless PTY fixtures | `…/tui/mouse/tests/headless` |
| Dashboard stage rows / Run chips / loading (this tree) | `BuildDashboardHitmap` + product resolve |
| CPR → known-origin chain smoke | `wrkcli/tui/tests/csi6n-origin` |

No PTY, no Bubble Tea program, no CSI 6n I/O here.

**Bug sealed here:** top-anchored geometry (UI near top, blank below) must map
a click on **gen-commit-msg** `[ Run ]` → `runStage == "gen-commit-msg"`,
**not** `"tag-next"`.

## DSN (Domain Specific Notion)

### Participants

- **Hitmap** — list of rectangular hit regions in **view-local** coordinates
  (Y = 0 at the first rendered dashboard line). Each region is a **Hit**:
  `y0,y1,x0,x1`, optional keyboard `focus` index, optional `runStage` id.
- **Stage row** — one dashboard line for a stage (e.g. `gen-commit-msg`,
  `tag-next`). Split into **left** (focus/toggle) and **right** (Run chip)
  regions; Run carries non-empty `runStage`.
- **Terminal frame** — absolute mouse coordinates (`absX`, `absY`, 0-based,
  Bubble Tea style), terminal `height` (rows), and `viewLines` (last render
  line count). Hitmap local Y lives in `0 .. viewLines-1`.
- **Origin** — row where the UI starts in the terminal:
  - **top-anchored**: origin ≈ 0 (screenshot layout: UI near top, blank below)
  - **bottom-anchored**: origin ≈ `height - viewLines` (inline paint at bottom)
  - **known origin**: caller supplies `originY`; single map `localY = absY - originY`
- **Resolver** — product `ResolveMouseHit` over dashboard hitmap (go-pkgs
  `mouse.Resolve` under the hood).
- **Loading gate** — when a stage/batch op is running, mouse clicks must not
  resolve to a run/focus action (miss / ignore).

### Behaviors

**Hitmap build (dashboard layout rules)**

- After layout at a given width: stage rows register left + Run hits (Run
  omitted when that stage’s Run is disabled, e.g. add-changes with no dirt).
- `gen-commit-msg` Run and `tag-next` Run have **distinct** local Y and
  correct `runStage` ids.

**Mouse resolve (product + dual-origin)**

- Inputs: `absX`, `absY`, `height`, `viewLines`, optional known `originY`,
  hitmap from `BuildDashboardHitmap`, optional loading flag.
- Known origin → `localY = absY - originY` (single candidate).
- Unknown origin → dual-origin candidates (top then bottom); prefer non-empty
  `runStage` when candidates disagree on the Run column.
- Loading → miss.
- Dual leaves assert **stage Run chips** via real dashboard geometry — not
  synthetic-only dual math (that lives in go-pkgs).

**Bug regression**

- Top-anchored: `height >> viewLines`, click gen-commit-msg Run cell →
  `runStage == "gen-commit-msg"` (must not be `"tag-next"`).
- Bottom-anchored: same relative cell still → `gen-commit-msg`.

## Decision Tree

```
mouse-resolve
├── dual-origin                      [origin unknown → try top + bottom; dashboard hitmap]
│   ├── top-anchored                 [height >> viewLines; UI at row 0]
│   │   ├── gen-commit-msg-run       BUG: Run click → gen-commit-msg not tag-next
│   │   ├── tag-next-run             distinct After-stage Run
│   │   └── blank-below-miss         click blank under UI → miss
│   └── bottom-anchored              [origin = height - viewLines]
│       ├── gen-commit-msg-run       same relative Run → gen-commit-msg
│       └── tag-next-run             same relative Run → tag-next
├── known-origin                     [explicit originY; single map]
│   ├── gen-commit-msg-run
│   └── tag-next-run
├── hitmap                           [BuildDashboardHitmap layout]
│   ├── distinct-run-stages          gen-commit-msg vs tag-next Y + ids
│   ├── run-column-vs-left           left focus vs right runStage
│   └── disabled-add-changes-run     disabled Run → no runStage hit
└── loading-ignore                   [loading gate]
    └── run-click-while-loading      Run click while loading → miss
```

## Test Index

| Leaf | Op | Description |
|------|-----|-------------|
| `dual-origin/top-anchored/gen-commit-msg-run` | resolve | **Bug regression:** top-anchored gen-commit-msg Run → `gen-commit-msg`, not `tag-next` |
| `dual-origin/top-anchored/tag-next-run` | resolve | Top-anchored tag-next Run → `tag-next` |
| `dual-origin/top-anchored/blank-below-miss` | resolve | Click blank rows below UI → miss |
| `dual-origin/bottom-anchored/gen-commit-msg-run` | resolve | Bottom-anchored gen-commit-msg Run → `gen-commit-msg` |
| `dual-origin/bottom-anchored/tag-next-run` | resolve | Bottom-anchored tag-next Run → `tag-next` |
| `known-origin/gen-commit-msg-run` | resolve | Known originY maps gen-commit-msg Run correctly |
| `known-origin/tag-next-run` | resolve | Known originY maps tag-next Run correctly |
| `hitmap/distinct-run-stages` | hitmap | gen-commit-msg and tag-next Run hits differ in Y and runStage |
| `hitmap/run-column-vs-left` | hitmap | gen-commit-msg left = focus, right = runStage |
| `hitmap/disabled-add-changes-run` | hitmap | add-changes disabled → no runStage region |
| `loading-ignore/run-click-while-loading` | resolve | Loading true → miss even on Run cell |

## How to Run

```sh
doctest vet ./wrkcli/tui/tests/mouse-resolve/
doctest test ./wrkcli/tui/tests/mouse-resolve/

# pure algorithm (go-pkgs)
cd external/dot-pkgs-master-2026-07-18-1/go-pkgs
doctest test ./tui/mouse/tests/hit-test/...
doctest test ./tui/mouse/tests/...
```

### Product API (`wrkcli/tui`)

```text
type Hit struct {
    Y0, Y1   int
    X0, X1   int
    Focus    int    // keyboard focus index; -1 if unused
    RunStage string // non-empty for Run chip
}

type BuildDashboardHitmapOpts struct {
    Width       int  // terminal content width used by layout
    AddDisabled bool // when true, add-changes Run is disabled
}

func BuildDashboardHitmap(opts BuildDashboardHitmapOpts) (hits []Hit, viewLines int)

type ResolveMouseHitOpts struct {
    AbsX, AbsY  int
    Height      int  // terminal rows
    ViewLines   int  // last render line count
    OriginY     *int // nil => dual-origin; non-nil => single map
    Hitmap      []Hit
    Loading     bool // true => always miss
}

type ResolveMouseHitResult struct {
    OK       bool
    Hit      Hit
    LocalY   int    // local Y used for the successful candidate (or -1)
    OriginKind string // "known" | "top" | "bottom" | "" when miss
}

func ResolveMouseHit(opts ResolveMouseHitOpts) ResolveMouseHitResult
```

Semantics:

- Known origin: `localY = absY - *OriginY`; hit-test once.
- Dual-origin (OriginY nil): evaluate **top** then **bottom** candidates;
  prefer non-empty `RunStage` when candidates hit different stages and the
  click X lands in a Run column; clamp bottom origin to ≥ 0.
- Hit-test: `y0 <= localY < y1` and (if `x1 > x0`) `x0 <= absX < x1`.
- Underlying pure resolve: go-pkgs `mouse.Resolve`.

```go
import (
	"fmt"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/wrk/wrkcli/tui"
)

// Request drives pure hitmap build and/or mouse resolve.
// Op: "resolve" (default) | "hitmap".
type Request struct {
	Op string

	// Layout
	Width       int
	AddDisabled bool

	// Geometry for resolve
	AbsX, AbsY int
	Height     int
	// ViewLines: 0 means use BuildDashboardHitmap's viewLines.
	ViewLines int
	// OriginYSet + OriginY: when OriginYSet, pass known origin to resolve.
	OriginYSet bool
	OriginY    int
	Loading    bool

	// Leaf helpers (not read by product code): which stage / target to aim.
	// StageID: "gen-commit-msg" | "tag-next" | "add-changes" | ...
	// Target: "run" | "left" | "blank-below"
	StageID string
	Target  string
	// ExtraBlank: rows of blank below UI for top-anchored geometry
	// (Height = viewLines + ExtraBlank when Height left 0).
	ExtraBlank int
	// OriginOffset: for bottom-anchored / known-origin, blank rows above UI
	// (origin = OriginOffset; Height = OriginOffset + viewLines when Height 0).
	OriginOffset int
}

type Response struct {
	// Hitmap (always filled when build succeeds)
	Hits      []tui.Hit
	ViewLines int

	// Resolve result (Op == "resolve")
	OK         bool
	Hit        tui.Hit
	LocalY     int
	OriginKind string
	RunStage   string
	Focus      int

	// Convenience: local coords of the aimed stage target (for debugging)
	AimedLocalY int
	AimedAbsY   int
	AimedAbsX   int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Width <= 0 {
		req.Width = 80
	}
	hits, viewLines := tui.BuildDashboardHitmap(tui.BuildDashboardHitmapOpts{
		Width:       req.Width,
		AddDisabled: req.AddDisabled,
	})
	if viewLines <= 0 {
		return nil, fmt.Errorf("BuildDashboardHitmap: viewLines=%d (expected > 0)", viewLines)
	}
	if req.ViewLines > 0 {
		viewLines = req.ViewLines
	}

	resp := &Response{
		Hits:      hits,
		ViewLines: viewLines,
		Focus:     -1,
		LocalY:    -1,
	}

	if req.Op == "hitmap" {
		return resp, nil
	}

	// Default Op: resolve — compute abs coords from StageID/Target when needed.
	absX, absY, aimedLocalY, err := aimClick(req, hits, viewLines)
	if err != nil {
		return nil, err
	}
	resp.AimedAbsX = absX
	resp.AimedAbsY = absY
	resp.AimedLocalY = aimedLocalY

	height := req.Height
	if height <= 0 {
		if req.OriginOffset > 0 {
			height = req.OriginOffset + viewLines
		} else if req.ExtraBlank > 0 {
			height = viewLines + req.ExtraBlank
		} else {
			height = viewLines + 20
		}
	}

	opts := tui.ResolveMouseHitOpts{
		AbsX:      absX,
		AbsY:      absY,
		Height:    height,
		ViewLines: viewLines,
		Hitmap:    hits,
		Loading:   req.Loading,
	}
	if req.OriginYSet {
		oy := req.OriginY
		opts.OriginY = &oy
	}

	got := tui.ResolveMouseHit(opts)
	resp.OK = got.OK
	resp.Hit = got.Hit
	resp.LocalY = got.LocalY
	resp.OriginKind = got.OriginKind
	if got.OK {
		resp.RunStage = got.Hit.RunStage
		resp.Focus = got.Hit.Focus
	}
	return resp, nil
}

// aimClick derives absolute click coords from StageID/Target and geometry mode.
// Leaves may also set AbsX/AbsY explicitly (AbsY >= 0 and Target empty uses them).
func aimClick(req *Request, hits []tui.Hit, viewLines int) (absX, absY, localY int, err error) {
	// Explicit coords: Target blank-below or both Abs set without StageID.
	if req.Target == "blank-below" {
		// First blank row under the UI for top-anchored paint (absY = viewLines).
		localY = viewLines // outside hitmap
		absY = viewLines
		if req.AbsX > 0 {
			absX = req.AbsX
		} else {
			absX = req.Width / 2
			if absX < 1 {
				absX = 40
			}
		}
		if req.AbsY > 0 {
			absY = req.AbsY
		}
		return absX, absY, localY, nil
	}

	if req.StageID == "" {
		// Fully explicit resolve without stage aim.
		return req.AbsX, req.AbsY, -1, nil
	}

	h, ok := findStageHit(hits, req.StageID, req.Target)
	if !ok {
		return 0, 0, -1, fmt.Errorf("no hitmap region for stage=%q target=%q (hits=%d)",
			req.StageID, req.Target, len(hits))
	}
	localY = h.Y0
	absX = (h.X0 + h.X1) / 2
	if absX < h.X0 {
		absX = h.X0
	}

	// Geometry → absolute Y
	switch {
	case req.OriginYSet:
		// Known origin: absY = originY + localY
		absY = req.OriginY + localY
	case req.OriginOffset > 0:
		// Bottom-anchored paint: origin = OriginOffset
		absY = req.OriginOffset + localY
	default:
		// Top-anchored paint: origin 0
		absY = localY
	}
	// Allow leaf override of AbsX/AbsY after aim (non-zero AbsY overrides).
	if req.AbsX > 0 {
		absX = req.AbsX
	}
	if req.AbsY > 0 {
		absY = req.AbsY
	}
	return absX, absY, localY, nil
}

func findStageHit(hits []tui.Hit, stageID, target string) (tui.Hit, bool) {
	wantRun := target == "" || target == "run"
	for _, h := range hits {
		if wantRun {
			if h.RunStage == stageID {
				return h, true
			}
			continue
		}
		// left: same row as runStage for this id, but focus region (empty runStage)
		if h.RunStage != "" {
			continue
		}
		// Match left by sharing Y with the stage's run hit.
		run, ok := findStageHit(hits, stageID, "run")
		if !ok {
			return tui.Hit{}, false
		}
		if h.Y0 == run.Y0 && h.Y1 == run.Y1 && h.X1 <= run.X0 {
			return h, true
		}
	}
	// left without run (e.g. disabled): scan any focus hit — not required for P1
	return tui.Hit{}, false
}
```
