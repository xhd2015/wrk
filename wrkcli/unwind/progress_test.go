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

func TestFormatProgressElapsed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "0ms"},
		{850 * time.Millisecond, "850ms"},
		{4800 * time.Millisecond, "4.8s"},
		{62 * time.Second, "1m02s"},
	}
	for _, tc := range cases {
		if got := formatProgressElapsed(tc.d); got != tc.want {
			t.Fatalf("formatProgressElapsed(%v)=%q want %q", tc.d, got, tc.want)
		}
	}
}

func TestFormatRowLineElapsedAfterOneline(t *testing.T) {
	t.Parallel()
	p := newActionProgress(progressConfig{W: &bytes.Buffer{}, Color: false, Indent: "      "}, []*Action{
		{ID: "a", Mode: ModePush, Lane: ".", Subject: Subject{Display: "."}},
	})
	row := p.rows["a"]
	row.Status = progDone
	row.StartedAt = time.Now().Add(-1500 * time.Millisecond)
	row.Elapsed = 1500 * time.Millisecond
	row.Oneline = "pushed master → origin/master"

	line := p.formatRowLineLocked(row)
	if !strings.Contains(line, "push") {
		t.Fatalf("missing label: %q", line)
	}
	onelineIdx := strings.Index(line, "pushed master")
	elapsedIdx := strings.Index(line, "1.5s")
	if onelineIdx < 0 || elapsedIdx < 0 {
		t.Fatalf("want oneline then elapsed: %q", line)
	}
	if elapsedIdx < onelineIdx {
		t.Fatalf("elapsed must follow oneline: %q", line)
	}
}
