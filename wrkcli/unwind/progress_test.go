package unwind

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestActionSinkRollsOneline(t *testing.T) {
	actions := []*Action{
		{ID: "a", Mode: ModeGenCommitMsg, Lane: "leaf", Subject: Subject{Display: "leaf"}},
		{ID: "b", Mode: ModeCommit, Lane: "leaf", Subject: Subject{Display: "leaf"}},
		{ID: "c", Mode: ModeAddAll, Lane: "root", Subject: Subject{Display: "root"}},
	}
	var buf bytes.Buffer
	p := newActionProgress(progressConfig{
		W:      &buf,
		Color:  false,
		Indent: "      ",
		LaneDisplay: func(lane string) string {
			if lane == "leaf" {
				return "external/leaf"
			}
			return "."
		},
	}, actions)
	// Non-TTY buffer → append-only; still exercise sink + grouping helpers.
	if len(p.groups) != 2 {
		t.Fatalf("groups=%d want 2", len(p.groups))
	}
	if p.groups[0].Display != "external/leaf" || p.groups[1].Display != "." {
		t.Fatalf("group order: %+v", p.groups)
	}

	p.Begin(2)
	p.Start("a")
	sink := p.ActionSink("a")
	_, _ = sink.Write([]byte("$ git diff --cached\n"))
	_, _ = sink.Write([]byte("Passing diff to agent...\n"))
	_, _ = sink.Write([]byte("Running agent..."))
	time.Sleep(10 * time.Millisecond)
	p.Finish("a", nil)

	p.mu.Lock()
	oneline := p.rows["a"].Oneline
	status := p.rows["a"].Status
	p.mu.Unlock()
	if status != progDone {
		t.Fatalf("status=%s", status)
	}
	if oneline != "Running agent..." {
		t.Fatalf("oneline=%q", oneline)
	}

	out := buf.String()
	if !strings.Contains(out, "gen-commit-msg") {
		t.Fatalf("missing action label in append log:\n%s", out)
	}
	p.Close()
}

func TestHostIODefaults(t *testing.T) {
	var io HostIO
	if io.Out() == nil || io.Err() == nil {
		t.Fatal("nil defaults")
	}
	var buf bytes.Buffer
	io = HostIO{Stdout: &buf, Stderr: &buf}
	if io.Out() != &buf || io.Err() != &buf {
		t.Fatal("custom writers not used")
	}
}
