package unwind

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor"
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

func TestStripProgressControls(t *testing.T) {
	t.Parallel()
	got := stripProgressControls("\x1b[5Ahello\x1b]10;?\x07 world\x1b[6n")
	if got != "hello world" {
		t.Fatalf("stripProgressControls=%q", got)
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("ESC leftover: %q", got)
	}
}

func TestClampProgressLine(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", 200)
	got := clampProgressLine(long, 40)
	if progressVisibleWidth(got) > 40 {
		t.Fatalf("width %d > 40: %q", progressVisibleWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis: %q", got)
	}
	colored := "\x1b[90m" + long + "\x1b[0m"
	got = clampProgressLine(colored, 40)
	if progressVisibleWidth(got) > 40 {
		t.Fatalf("colored width %d > 40: %q", progressVisibleWidth(got), got)
	}
}

func TestProgressFrameWidthXenl(t *testing.T) {
	t.Parallel()
	if got := progressFrameWidth(80); got != 79 {
		t.Fatalf("progressFrameWidth(80)=%d want 79", got)
	}
	if got := progressFrameWidth(1); got != 1 {
		t.Fatalf("progressFrameWidth(1)=%d want 1", got)
	}
}

func TestBuildProgressFrameClearsLeftover(t *testing.T) {
	t.Parallel()
	frame := buildProgressFrame(4, []string{"a", "b"}, 40)
	if !strings.HasPrefix(frame, cursorHide) || !strings.HasSuffix(frame, cursorShow) {
		t.Fatalf("hide/show: %q", frame)
	}
	if !strings.Contains(frame, cursor.Up(4)) {
		t.Fatalf("missing Up(old): %q", frame)
	}
	if !strings.Contains(frame, cursor.Up(2)) {
		t.Fatalf("missing Up(extra): %q", frame)
	}
	if strings.Count(frame, cursor.ClearLine) != 4 {
		t.Fatalf("ClearLine count=%d want 4: %q", strings.Count(frame, cursor.ClearLine), frame)
	}
}

func TestSetOnelineStripsCursorCSI(t *testing.T) {
	t.Parallel()
	p := newActionProgress(progressConfig{W: &bytes.Buffer{}, Color: false, Indent: "      "}, []*Action{
		{ID: "a", Mode: ModeAddAll, Lane: "leaf", Subject: Subject{Display: "leaf"}},
	})
	p.setOneline("a", "\x1b[5A"+strings.Repeat("x", 8)+"\x1b]11;?\x07")
	p.mu.Lock()
	oneline := p.rows["a"].Oneline
	p.mu.Unlock()
	if strings.Contains(oneline, "\x1b") {
		t.Fatalf("oneline kept ESC: %q", oneline)
	}
	if !strings.Contains(oneline, "xxxxxxxx") {
		t.Fatalf("oneline=%q", oneline)
	}
}

func TestRedrawLastFrameOneHeader(t *testing.T) {
	var buf bytes.Buffer
	p := newActionProgress(progressConfig{
		W:      &buf,
		Color:  false,
		Indent: "      ",
		LaneDisplay: func(lane string) string {
			if lane == "leaf" {
				return "external/dot-pkgs"
			}
			return "."
		},
	}, []*Action{
		{ID: "add", Mode: ModeAddAll, Lane: "leaf", Subject: Subject{Display: "leaf"}},
		{ID: "pin", Mode: ModePin, Lane: "leaf", Subject: Subject{Display: "leaf"}, Detail: "<- github.com/xhd2015/dot-pkgs/go-pkgs@v0.0.174"},
		{ID: "root", Mode: ModeAddAll, Lane: "root", Subject: Subject{Display: "root"}},
	})
	p.block = true
	p.termWidth = 40
	p.Begin(2)
	p.Start("add")
	p.setOneline("add", "\x1b[5A"+strings.Repeat("y", 80))
	p.Finish("add", nil)
	p.Start("pin")
	p.Finish("pin", nil)
	p.Close()

	frame := lastProgressFrame(buf.String())
	plain := progressFramePlain(frame)
	if n := strings.Count(plain, "external/dot-pkgs"); n != 1 {
		t.Fatalf("header count=%d in last frame:\n%s", n, plain)
	}
	if strings.Contains(p.rows["add"].Oneline, "\x1b") {
		t.Fatalf("stored oneline has ESC: %q", p.rows["add"].Oneline)
	}
	for _, line := range strings.Split(plain, "\n") {
		if line == "" {
			continue
		}
		if w := progressVisibleWidth(line); w > progressFrameWidth(40) {
			t.Fatalf("wrapped line width %d: %q", w, line)
		}
	}
}

func lastProgressFrame(s string) string {
	i := strings.LastIndex(s, cursorHide)
	if i < 0 {
		return s
	}
	rest := s[i+len(cursorHide):]
	if j := strings.LastIndex(rest, cursorShow); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

func progressFramePlain(frame string) string {
	frame = progressCSIPattern.ReplaceAllString(frame, "")
	frame = strings.ReplaceAll(frame, "\r", "")
	return frame
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
