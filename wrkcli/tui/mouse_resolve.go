package tui

import (
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
)

// Hit is a rectangular mouse hit region in view-local coordinates
// (Y = 0 at the first rendered dashboard line).
type Hit struct {
	Y0, Y1   int
	X0, X1   int
	Focus    int    // keyboard focus index; -1 if unused
	RunStage string // non-empty for Run chip
}

// BuildDashboardHitmapOpts configures pure hitmap layout.
type BuildDashboardHitmapOpts struct {
	Width       int
	AddDisabled bool
}

// BuildDashboardHitmap returns view-local hit regions and the rendered line count.
func BuildDashboardHitmap(opts BuildDashboardHitmapOpts) (hits []Hit, viewLines int) {
	m := newTeaDashModel(RunDashboardOpts{})
	if opts.Width > 0 {
		m.width = opts.Width
	} else {
		m.width = 80
	}
	m.addDisabled = opts.AddDisabled
	m.addAll = !opts.AddDisabled
	m.color = false
	_ = m.renderView()
	hits = dashHitsToHits(m.hitmap)
	return hits, m.viewLines
}

// ResolveMouseHitOpts is pure input for absolute mouse → hit resolution.
type ResolveMouseHitOpts struct {
	AbsX, AbsY int
	Height     int
	ViewLines  int
	OriginY    *int
	Hitmap     []Hit
	Loading    bool
}

// ResolveMouseHitResult is the outcome of ResolveMouseHit.
type ResolveMouseHitResult struct {
	OK         bool
	Hit        Hit
	LocalY     int
	OriginKind string
}

// ResolveMouseHit maps an absolute terminal click onto a view-local Hit.
func ResolveMouseHit(opts ResolveMouseHitOpts) ResolveMouseHitResult {
	if opts.Loading {
		return ResolveMouseHitResult{OK: false, LocalY: -1}
	}
	r := mouse.Resolve(mouse.ResolveOpts{
		AbsX: opts.AbsX, AbsY: opts.AbsY,
		Height: opts.Height, ViewLines: opts.ViewLines,
		OriginY: opts.OriginY,
		Hits:    hitsToMouse(opts.Hitmap),
	})
	if !r.OK {
		return ResolveMouseHitResult{OK: false, LocalY: -1}
	}
	return ResolveMouseHitResult{
		OK: true, Hit: mouseHitToDash(r.Hit, opts.Hitmap),
		LocalY: r.LocalY, OriginKind: r.Kind,
	}
}

// ParseCPR re-exports stream CPR parse for doctests.
func ParseCPR(buf []byte) (row1, col1 int, ok bool) {
	return mouse.ParseCPR(buf)
}

// OriginFromCPR keeps clamped semantics for sealed doctests.
// Live origin tracking uses mouse.Tracker (strict OriginFromCPR).
func OriginFromCPR(row1, viewLines int) (originY0 int, ok bool) {
	return mouse.OriginFromCPRClamped(row1, viewLines)
}

func hitsToMouse(in []Hit) []mouse.Hit {
	if len(in) == 0 {
		return nil
	}
	out := make([]mouse.Hit, len(in))
	for i, h := range in {
		// RunStage as ID so dual-origin prefers Run chips over focus-only (empty ID).
		out[i] = mouse.Hit{Y0: h.Y0, Y1: h.Y1, X0: h.X0, X1: h.X1, ID: h.RunStage}
	}
	return out
}

func mouseHitToDash(mh mouse.Hit, original []Hit) Hit {
	for _, h := range original {
		if h.Y0 == mh.Y0 && h.Y1 == mh.Y1 && h.X0 == mh.X0 && h.X1 == mh.X1 {
			return h
		}
	}
	return Hit{Y0: mh.Y0, Y1: mh.Y1, X0: mh.X0, X1: mh.X1, Focus: -1, RunStage: mh.ID}
}

func dashHitsToHits(in []dashHit) []Hit {
	if len(in) == 0 {
		return nil
	}
	out := make([]Hit, len(in))
	for i, h := range in {
		out[i] = Hit{
			Y0: h.y0, Y1: h.y1, X0: h.x0, X1: h.x1,
			Focus: h.focus, RunStage: h.runStage,
		}
	}
	return out
}

func dashHitsToMouse(in []dashHit) []mouse.Hit {
	return hitsToMouse(dashHitsToHits(in))
}
