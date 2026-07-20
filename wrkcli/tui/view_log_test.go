package tui

import (
	"strings"
	"testing"
)

// countLogBodyLines returns how many body lines sit between the Log section
// header and the footer hint line.
func countLogBodyLines(out string) int {
	lines := strings.Split(out, "\n")
	start, end := -1, -1
	for i, ln := range lines {
		// Section HR contains "─ Log " or bold-cyan Log with dashes nearby.
		if strings.Contains(ln, "Log") && strings.Contains(ln, "─") && start < 0 {
			// Prefer section header (├ or ANSI) over log content "stage │".
			if strings.Contains(ln, "├") || strings.Contains(ln, "\x1b") && !strings.Contains(ln, "│ ") {
				start = i
				continue
			}
			if strings.Contains(ln, "─ Log") || strings.Contains(ln, " Log ") {
				start = i
				continue
			}
		}
		if start >= 0 && strings.Contains(ln, "↑↓ row") {
			end = i
			break
		}
	}
	if start < 0 || end < 0 || end <= start {
		return -1
	}
	return end - start - 1
}

func TestRenderViewLogSectionEmpty(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{
		WorkDir: "/tmp",
		Status:  "ready",
	})
	m.width = 100
	m.color = false

	out := m.renderView()
	if !strings.Contains(out, "Log") {
		t.Fatalf("expected Log section header in view:\n%s", out)
	}
	if !strings.Contains(out, "(no log yet)") {
		t.Fatalf("expected empty log placeholder:\n%s", out)
	}
	n := countLogBodyLines(out)
	if n != dashLogViewLines {
		t.Fatalf("empty log body lines=%d want %d\n%s", n, dashLogViewLines, out)
	}
	lines := strings.Split(out, "\n")
	if m.viewLines != len(lines) {
		t.Fatalf("viewLines=%d len(split)=%d", m.viewLines, len(lines))
	}
}

func TestRenderViewLogSectionShowsLines(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{
		WorkDir: "/tmp",
		Status:  "running…",
	})
	m.width = 100
	m.color = false
	m.loadingID = "sync"
	m.appendLog(LogLine{Stage: "sync", Text: "fetching remotes"})
	m.appendLog(LogLine{Stage: "sync", Text: "unique-log-token-xyz"})

	out := m.renderView()
	if !strings.Contains(out, "Log") {
		t.Fatalf("expected Log section:\n%s", out)
	}
	if !strings.Contains(out, "unique-log-token-xyz") {
		t.Fatalf("expected log text in view:\n%s", out)
	}
	if !strings.Contains(out, "sync │") {
		t.Fatalf("expected stage prefix in log line:\n%s", out)
	}
	if strings.Contains(out, "(no log yet)") {
		t.Fatalf("should not show empty placeholder when logs present:\n%s", out)
	}
	n := countLogBodyLines(out)
	if n != dashLogViewLines {
		t.Fatalf("log body lines=%d want %d\n%s", n, dashLogViewLines, out)
	}

	logIdx := strings.Index(out, "─ Log ")
	tokenIdx := strings.Index(out, "unique-log-token-xyz")
	if logIdx < 0 || tokenIdx < logIdx {
		t.Fatalf("log text should appear under Log section; logIdx=%d tokenIdx=%d", logIdx, tokenIdx)
	}

	lines := strings.Split(out, "\n")
	if m.viewLines != len(lines) {
		t.Fatalf("viewLines=%d len(split)=%d", m.viewLines, len(lines))
	}
}

func TestRenderViewLogSectionShowsLastN(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	for i := 0; i < dashLogViewLines+3; i++ {
		m.appendLog(LogLine{Stage: "run-all", Text: "row-" + itoa(i)})
	}
	out := m.renderView()
	if strings.Contains(out, "row-0") {
		t.Fatalf("expected oldest rows truncated from view:\n%s", out)
	}
	last := "row-" + itoa(dashLogViewLines+2)
	if !strings.Contains(out, last) {
		t.Fatalf("expected newest row %q in view:\n%s", last, out)
	}
	n := countLogBodyLines(out)
	if n != dashLogViewLines {
		t.Fatalf("log body lines=%d want %d\n%s", n, dashLogViewLines, out)
	}
}

func TestLogViewportAlwaysThreeAndStable(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false

	out0 := m.renderView()
	n0 := m.viewLines
	body0 := countLogBodyLines(out0)
	if body0 != 3 {
		t.Fatalf("empty: body=%d want 3\n%s", body0, out0)
	}

	m.appendLog(LogLine{Stage: "push", Text: "fatal: no upstream configured"})
	out1 := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced after 1 log: %d → %d", n0, m.viewLines)
	}
	if countLogBodyLines(out1) != 3 {
		t.Fatalf("1 log: body want 3\n%s", out1)
	}
	if !strings.Contains(out1, "no upstream") {
		t.Fatalf("expected log content:\n%s", out1)
	}
	if strings.Contains(out1, "(no log yet)") {
		t.Fatalf("placeholder should clear when logs present:\n%s", out1)
	}

	for i := 0; i < 10; i++ {
		m.appendLog(LogLine{Stage: "run-all", Text: "line-" + itoa(i)})
	}
	out2 := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced after many logs: %d → %d", n0, m.viewLines)
	}
	if countLogBodyLines(out2) != 3 {
		t.Fatalf("many logs: body want 3\n%s", out2)
	}
	// Newest at bottom of viewport.
	if !strings.Contains(out2, "line-9") {
		t.Fatalf("expected newest log in view:\n%s", out2)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
