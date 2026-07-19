# CSI 6n → known-origin resolve chain (`wrkcli/tui` product adapters)

## Version
0.0.2

**Plan phase P3 (slim):** product-only chain doctests for the inline dashboard
TUI in `github.com/xhd2015/wrk/wrkcli/tui`.

This tree seals **CPR bytes → originY → `ResolveMouseHit` with dashboard
hitmap** (gen-commit-msg Run). Pure algorithm contracts live in go-pkgs:

| Concern | Owner |
|---------|--------|
| `ParseCPR`, strict `OriginFromCPR`, dual-origin math, HitTest | `external/dot-pkgs-…/go-pkgs/tui/mouse/tests` (pure) |
| PTY / headless mouse fixtures | `…/tui/mouse/tests/headless` |
| Dashboard layout hitmap + stage Run chips | `wrkcli/tui/tests/mouse-resolve` |
| Thin product chain (this tree) | CPR-derived origin + `BuildDashboardHitmap` resolve |

**Dropped in P3 (do not re-add here):** pure `parse-cpr/*` and
`origin-from-cpr/*` leaves — owned by go-pkgs. Live origin reject rule is in
go-pkgs; wrk `OriginFromCPR` re-export may use clamped helper for legacy
adapters, but pure edges are not re-tested under wrk.

No PTY, no live iTerm2, no Bubble Tea program.

## DSN (Domain Specific Notion)

### Participants

- **CPR reply** — terminal answer form `ESC [ <row> ; <col> R` (1-based).
  Injected as bytes in tests (no I/O).
- **Product adapters** — `tui.ParseCPR` / `tui.OriginFromCPR` (thin re-exports
  of go-pkgs) plus `BuildDashboardHitmap` / `ResolveMouseHit`.
- **Dashboard hitmap** — view-local stage regions from real layout (same as
  mouse-resolve).
- **Known-origin resolve** — after successful CPR → originY0, click aimed at
  gen-commit-msg Run with `OriginY = &originY0`.

### Behaviors

**Product chain**

- Synthesize or inject CPR for “cursor on last paint line with blank rows
  above” → parse → origin → resolve gen-commit-msg Run with known origin.
- Empty / failed CPR → `OriginOK=false`; do not claim a known-origin resolve
  hit (live TUI leaves `OriginY` nil → dual-origin, covered under mouse-resolve).

## Decision Tree

```
csi6n-origin
└── chain-known-origin                 [Op=chain — parse→origin→resolve product]
    ├── cpr-to-gen-commit-msg-run      CPR → originY → gen-commit-msg Run hit
    └── cpr-fail-no-origin             bad CPR → !OriginOK (no known origin)
```

## Test Index

| Leaf | Op | Description |
|------|-----|-------------|
| `chain-known-origin/cpr-to-gen-commit-msg-run` | chain | CPR-derived origin maps gen-commit-msg Run |
| `chain-known-origin/cpr-fail-no-origin` | chain | Failed CPR does not yield known origin |

## How to Run

```sh
# product chain (this tree)
doctest vet ./wrkcli/tui/tests/csi6n-origin/
doctest test ./wrkcli/tui/tests/csi6n-origin/

# pure algorithm (go-pkgs — not this tree)
cd external/dot-pkgs-master-2026-07-18-1/go-pkgs
doctest test ./tui/mouse/tests/hit-test/...
doctest test ./tui/mouse/tests/...
```

### Product API (adapters under `wrkcli/tui`)

```text
// Thin re-exports / dashboard wrappers used by chain Run:
func ParseCPR(buf []byte) (row1, col1 int, ok bool)
func OriginFromCPR(row1, viewLines int) (originY0 int, ok bool)
func BuildDashboardHitmap(opts BuildDashboardHitmapOpts) (hits []Hit, viewLines int)
func ResolveMouseHit(opts ResolveMouseHitOpts) ResolveMouseHitResult
```

Pure CPR/origin/dual-origin semantics: see go-pkgs `tui/mouse` doctests and
unit expansion matrix. Live I/O (write `ESC [6n`, read stdin timeout) is out of
scope here.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/wrk/wrkcli/tui"
)

// Request drives the product CPR → origin → ResolveMouseHit chain.
// Op: "chain" (only remaining operation after P3 slim).
type Request struct {
	Op string

	// Buf: raw CPR-like bytes. Empty + BlankAbove < 0 forces parse fail.
	// Empty + BlankAbove >= 0 synthesizes CPR for last-line cursor.
	Buf []byte

	// ViewLines: 0 means use BuildDashboardHitmap's viewLines.
	ViewLines int

	// BlankAbove: blank terminal rows above the UI when synthesizing CPR
	// (row1 = BlankAbove + viewLines). BlankAbove < 0 disables synthesis.
	BlankAbove int

	// Layout / aim (dashboard gen-commit-msg Run by default)
	Width   int
	StageID string
	Target  string
}

type Response struct {
	ParseOK bool
	Row1    int
	Col1    int

	OriginOK bool
	OriginY0 int

	ViewLines  int
	ResolveOK  bool
	RunStage   string
	OriginKind string
	LocalY     int
	AimedAbsY  int
	AimedAbsX  int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Op == "" {
		req.Op = "chain"
	}
	if req.Op != "chain" {
		return nil, fmt.Errorf("unknown Op %q (only chain after P3 slim)", req.Op)
	}
	resp := &Response{LocalY: -1}
	return runChain(t, req, resp)
}

func runChain(t *testing.T, req *Request, resp *Response) (*Response, error) {
	_ = t
	width := req.Width
	if width <= 0 {
		width = 80
	}
	hits, viewLines := tui.BuildDashboardHitmap(tui.BuildDashboardHitmapOpts{
		Width:       width,
		AddDisabled: false,
	})
	if viewLines <= 0 {
		return nil, fmt.Errorf("BuildDashboardHitmap: viewLines=%d", viewLines)
	}
	if req.ViewLines > 0 {
		viewLines = req.ViewLines
	}
	resp.ViewLines = viewLines

	buf := req.Buf
	if len(buf) == 0 && req.BlankAbove >= 0 && req.StageID != "" {
		// Synthesize CPR: cursor on last view line after paint.
		// row1 (1-based) = originY0 + viewLines = BlankAbove + viewLines
		row1 := req.BlankAbove + viewLines
		buf = []byte(fmt.Sprintf("\x1b[%d;1R", row1))
	}

	row1, col1, parseOK := tui.ParseCPR(buf)
	resp.ParseOK = parseOK
	resp.Row1 = row1
	resp.Col1 = col1
	if !parseOK {
		resp.OriginOK = false
		return resp, nil
	}

	oy, originOK := tui.OriginFromCPR(row1, viewLines)
	resp.OriginOK = originOK
	resp.OriginY0 = oy
	if !originOK {
		return resp, nil
	}

	stageID := req.StageID
	if stageID == "" {
		stageID = "gen-commit-msg"
	}
	target := req.Target
	if target == "" {
		target = "run"
	}
	h, ok := findStageHit(hits, stageID, target)
	if !ok {
		return nil, fmt.Errorf("no hitmap region for stage=%q target=%q", stageID, target)
	}
	localY := h.Y0
	absX := (h.X0 + h.X1) / 2
	if absX < h.X0 {
		absX = h.X0
	}
	absY := oy + localY
	resp.AimedAbsX = absX
	resp.AimedAbsY = absY

	originY := oy
	got := tui.ResolveMouseHit(tui.ResolveMouseHitOpts{
		AbsX:      absX,
		AbsY:      absY,
		Height:    oy + viewLines + 5,
		ViewLines: viewLines,
		OriginY:   &originY,
		Hitmap:    hits,
		Loading:   false,
	})
	resp.ResolveOK = got.OK
	resp.LocalY = got.LocalY
	resp.OriginKind = got.OriginKind
	if got.OK {
		resp.RunStage = got.Hit.RunStage
	}
	return resp, nil
}

func findStageHit(hits []tui.Hit, stageID, target string) (tui.Hit, bool) {
	wantRun := target == "" || target == "run"
	for _, h := range hits {
		if wantRun {
			if h.RunStage == stageID {
				return h, true
			}
		}
	}
	return tui.Hit{}, false
}
```
