package wrkcli

import (
	"bytes"
	"strings"
	"testing"
)

func TestShipProgressAppendEmitsOnceOnFinish(t *testing.T) {
	var buf bytes.Buffer
	p := &shipProgress{
		w:         &buf,
		block:     false, // append mode
		indent:    "      ",
		rows:      map[string]*shipProgRow{},
		sinks:     map[string]*shipProgSink{},
		captures:  map[string]*bytes.Buffer{},
		termWidth: 80,
		stopSpin:  make(chan struct{}),
	}
	for _, id := range []string{"tag-next+push", "sync"} {
		p.order = append(p.order, id)
		p.rows[id] = &shipProgRow{ID: id, Label: id, Status: shipProgWaiting}
		cap := &bytes.Buffer{}
		p.captures[id] = cap
		p.sinks[id] = &shipProgSink{p: p, id: id, capture: cap}
	}

	p.Begin()
	if buf.Len() != 0 {
		t.Fatalf("Begin must stay silent in append mode; got %q", buf.String())
	}

	p.Start("tag-next+push")
	p.setOneline("tag-next+push", "partial")
	if buf.Len() != 0 {
		t.Fatalf("Start/oneline must not append; got %q", buf.String())
	}

	p.Finish("tag-next+push", "tagged v0.0.12 · pushed", nil)
	p.Finish("tag-next+push", "tagged v0.0.12 · pushed", nil) // idempotent
	p.Finish("sync", "synced: 0 into main, 0 into worktrees, 0 skipped", nil)
	p.Close()

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("want exactly 2 final rows; got %d: %q", len(lines), out)
	}
	if !strings.Contains(lines[0], "tag-next+push") || !strings.Contains(lines[0], "tagged v0.0.12") {
		t.Fatalf("row0: %q", lines[0])
	}
	if !strings.Contains(lines[1], "sync") || !strings.Contains(lines[1], "synced:") {
		t.Fatalf("row1: %q", lines[1])
	}
	if strings.Contains(out, "partial") {
		t.Fatalf("must not keep Start oneline as final; got %q", out)
	}
}

func TestClampShipProgLine(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := clampShipProgLine(long, 40)
	if shipVisibleWidth(got) > 40 {
		t.Fatalf("width %d > 40: %q", shipVisibleWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis: %q", got)
	}
}

func TestShipReinstallDiagLines(t *testing.T) {
	capture := strings.Join([]string{
		"notice: bin wrk: preferring ./script/wrk over ./cmd/wrk",
		"go install ./cmd/foo",
		"skip: bar (not in /tmp/bin)",
		"warning: bin baz: ambiguous under cmd (./cmd/baz, ./cmd/baz2); skipping",
		"warning: reinstall finished with 1 failed",
		"reinstalled 1, skipped 1, failed 1",
	}, "\n")
	got := shipReinstallDiagLines(capture)
	want := []string{
		"notice: bin wrk: preferring ./script/wrk over ./cmd/wrk",
		"warning: bin baz: ambiguous under cmd (./cmd/baz, ./cmd/baz2); skipping",
		"warning: reinstall finished with 1 failed",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines %q; want %d %q", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestFlushShipReinstallDiags(t *testing.T) {
	var errBuf bytes.Buffer
	flushShipReinstallDiags(&errBuf, "go install ./cmd/foo\nnotice: bin wrk: preferring ./script/wrk over ./cmd/wrk\n")
	got := errBuf.String()
	if !strings.Contains(got, "notice: bin wrk:") {
		t.Fatalf("expected notice in errW; got %q", got)
	}
	if strings.Contains(got, "go install") {
		t.Fatalf("must not flush install noise; got %q", got)
	}

	errBuf.Reset()
	flushShipReinstallDiags(&errBuf, "go install ./cmd/foo\nreinstalled 1, skipped 0, failed 0\n")
	if errBuf.Len() != 0 {
		t.Fatalf("no diags → silent; got %q", errBuf.String())
	}
}

func TestIsShipReinstallDiagLineANSI(t *testing.T) {
	colored := ansiOrange + "warning:" + ansiReset + " bin x: ambiguous under cmd (a, b); skipping"
	if !isShipReinstallDiagLine(colored) {
		t.Fatalf("ANSI warning: bin should match: %q", colored)
	}
	if isShipReinstallDiagLine("warning: skip master-x: dirty worktree") {
		t.Fatal("sync skip warning must not match reinstall diag filter")
	}
}
