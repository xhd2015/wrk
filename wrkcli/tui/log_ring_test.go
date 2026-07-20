package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRunDashboardOpStreamsLines(t *testing.T) {
	logCh := make(chan dashLogLineMsg, 16)
	opts := RunDashboardOpts{
		RunCompose: func(workDir string, r Recipe) error {
			fmt.Println("stream-line-one")
			fmt.Println("stream-line-two")
			return nil
		},
		ComposeArgv: func(r Recipe) []string { return []string{"wrk", "sync"} },
	}
	status, err := runDashboardOpCaptured(opts, "/tmp", false, Recipe{}, "sync", false, logCh, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if status == "" {
		t.Fatal("expected non-empty status")
	}
	// Reader finished before return; drain buffered lines.
	var texts []string
drain:
	for {
		select {
		case msg := <-logCh:
			texts = append(texts, msg.Text)
			if msg.Stage != "sync" {
				t.Fatalf("stage=%q want sync", msg.Stage)
			}
		default:
			break drain
		}
	}
	// Legacy bridge captures fmt from compose; structured LogFunc may also add status lines.
	joined := strings.Join(texts, "\n")
	if !strings.Contains(joined, "stream-line-one") || !strings.Contains(joined, "stream-line-two") {
		t.Fatalf("streamed lines=%v want stream-line-one/two", texts)
	}
}

func TestAppendLogRingTruncatesAtCap(t *testing.T) {
	m := &teaDashModel{}
	for i := 0; i < maxDashLogs+50; i++ {
		m.appendLog(LogLine{
			Stage: "sync",
			Text:  fmt.Sprintf("line-%d", i),
			At:    time.Unix(int64(i), 0),
		})
	}
	if got := len(m.logs); got != maxDashLogs {
		t.Fatalf("len(logs)=%d want %d", got, maxDashLogs)
	}
	// Oldest kept should be the 50th written (0..49 dropped).
	wantFirst := fmt.Sprintf("line-%d", 50)
	if m.logs[0].Text != wantFirst {
		t.Fatalf("first log text=%q want %q", m.logs[0].Text, wantFirst)
	}
	wantLast := fmt.Sprintf("line-%d", maxDashLogs+49)
	if m.logs[len(m.logs)-1].Text != wantLast {
		t.Fatalf("last log text=%q want %q", m.logs[len(m.logs)-1].Text, wantLast)
	}
}

func TestAppendLogPreservesStage(t *testing.T) {
	m := &teaDashModel{}
	m.appendLog(LogLine{Stage: "add-changes", Text: "staging"})
	if m.logs[0].Stage != "add-changes" || m.logs[0].Text != "staging" {
		t.Fatalf("unexpected log entry: %+v", m.logs[0])
	}
}

func TestDashLogLineMsgAppendsAndRearms(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	ch := make(chan dashLogLineMsg, 1)
	m.logCh = ch
	m.loadingID = "sync"

	next, cmd := m.Update(dashLogLineMsg{Stage: "sync", Text: "hello from op"})
	mm := next.(*teaDashModel)
	if len(mm.logs) != 1 || mm.logs[0].Text != "hello from op" {
		t.Fatalf("logs after msg: %+v", mm.logs)
	}
	if cmd == nil {
		t.Fatal("expected listenLogs re-arm cmd while logCh open")
	}

	// Closed channel → clear logCh.
	close(ch)
	next, _ = mm.Update(dashLogClosedMsg{})
	mm = next.(*teaDashModel)
	if mm.logCh != nil {
		t.Fatal("logCh should be nil after dashLogClosedMsg")
	}
}
