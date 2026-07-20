package tui

import (
	"strings"
	"testing"
)

func TestDashLayoutWidth_1_5x(t *testing.T) {
	if got := dashLayoutWidth(50); got != 84 {
		t.Fatalf("min: got %d want 84", got)
	}
	if got := dashLayoutWidth(90); got != 90 {
		t.Fatalf("mid: got %d want 90", got)
	}
	if got := dashLayoutWidth(200); got != 108 {
		t.Fatalf("max: got %d want 108", got)
	}
}

func TestInlinePreviewOnSameLineWithPrefix(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false

	out0 := m.renderView()
	n0 := m.viewLines

	m.applyPreview("add-changes", "3 files")
	out := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced: %d → %d", n0, m.viewLines)
	}
	if !strings.Contains(out, "preview: 3 files") {
		t.Fatalf("expected readable preview: prefix:\n%s", out)
	}
	// Meta on same line as operation — not a standalone secondary line.
	lines := strings.Split(out, "\n")
	addY := -1
	for i, ln := range lines {
		if strings.Contains(ln, "add changes") && strings.Contains(ln, "preview: 3 files") {
			addY = i
			break
		}
	}
	if addY < 0 {
		t.Fatalf("expected preview on add-changes row:\n%s", out)
	}
	if addY+1 >= len(lines) || !strings.Contains(lines[addY+1], "gen-commit-msg") {
		t.Fatalf("line after add-changes should be gen-commit-msg, got %q\nbefore render:\n%s\nafter:\n%s",
			safeLine(lines, addY+1), out0, out)
	}
}

func safeLine(lines []string, i int) string {
	if i < 0 || i >= len(lines) {
		return "<oob>"
	}
	return lines[i]
}

func TestInlinePendingPreviewEllipsis(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{
		WorkDir: "/tmp",
		StagePreview: func(workDir, stageID string) StagePreviewResult {
			return StagePreviewResult{Preview: "14 files"}
		},
	})
	m.width = 100
	m.color = false
	if m.previewSettledFor("add-changes") {
		t.Fatal("expected unsettled before preview msgs")
	}
	out0 := m.renderView()
	n0 := m.viewLines
	if !strings.Contains(out0, "preview: …") {
		t.Fatalf("expected pending preview: …:\n%s", out0)
	}
	if strings.Contains(out0, "14 files") {
		t.Fatalf("should not show filled preview while pending:\n%s", out0)
	}

	m.applyPreview("add-changes", "14 files")
	out1 := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced: %d → %d", n0, m.viewLines)
	}
	for _, ln := range strings.Split(out1, "\n") {
		if strings.Contains(ln, "add changes") && strings.Contains(ln, "preview: 14 files") {
			return
		}
	}
	t.Fatalf("preview: 14 files not on add-changes row:\n%s", out1)
}

func TestViewLinesStableAfterAllPreviews(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{
		WorkDir: "/tmp",
		StagePreview: func(workDir, stageID string) StagePreviewResult {
			switch stageID {
			case "add-changes", "gen-commit-msg", "commit":
				return StagePreviewResult{Preview: "3 files"}
			case "merge-back", "done":
				return StagePreviewResult{Preview: "ahead 2"}
			case "push":
				return StagePreviewResult{Preview: "up to date"}
			default:
				return StagePreviewResult{}
			}
		},
	})
	m.width = 100
	m.color = false
	before := m.renderView()
	n0 := m.viewLines

	for _, id := range dashStageIDs {
		m.applyPreview(id, m.opts.StagePreview(m.workDir, id).Preview)
	}
	after := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced: before=%d after=%d\nbefore:\n%s\nafter:\n%s", n0, m.viewLines, before, after)
	}
	if !strings.Contains(after, "preview: 3 files") || !strings.Contains(after, "preview: ahead 2") {
		t.Fatalf("expected labeled previews:\n%s", after)
	}
}

func TestApplyPreviewEmptyOmitsMeta(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	m.applyPreview("push", "ahead 2")
	_ = m.renderView()
	n0 := m.viewLines
	m.applyPreview("push", "  ")
	out := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced: %d → %d", n0, m.viewLines)
	}
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "push") && strings.Contains(ln, "[ Run ]") {
			if strings.Contains(ln, "ahead 2") {
				t.Fatalf("empty settle should drop meta:\n%s", ln)
			}
		}
	}
}

func TestInlineMainAndAfterMeta(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	m.applyPreview("merge-back", "ahead 4")
	m.applyPreview("done", "ahead 4")
	m.applyPreview("push", "up to date")

	out := m.renderView()
	found := false
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "MERGE BACK") && strings.Contains(ln, "preview: ahead 4") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("MERGE BACK and preview: ahead 4 should share a line:\n%s", out)
	}
	if !strings.Contains(out, "preview: up to date") {
		t.Fatalf("expected push preview:\n%s", out)
	}
}

func TestMarkPreviewsPendingInline(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{
		WorkDir: "/tmp",
		StagePreview: func(workDir, stageID string) StagePreviewResult {
			return StagePreviewResult{Preview: "x"}
		},
	})
	m.width = 100
	m.color = false
	m.applyPreview("add-changes", "14 files")
	_ = m.renderView()
	n0 := m.viewLines
	m.markPreviewsPending()
	out := m.renderView()
	if m.viewLines != n0 {
		t.Fatalf("viewLines bounced: %d → %d", n0, m.viewLines)
	}
	if strings.Contains(out, "14 files") {
		t.Fatalf("pending should hide old preview:\n%s", out)
	}
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "add changes") && strings.Contains(ln, "preview: …") {
			return
		}
	}
	t.Fatalf("expected preview: … on add-changes row while pending:\n%s", out)
}

func TestRunStagePreviewSoftFailTimeout(t *testing.T) {
	if got := runStagePreview(nil, "/tmp", "add-changes"); got.Preview != "" || len(got.Logs) != 0 {
		t.Fatalf("nil fn → empty, got %+v", got)
	}
	got := runStagePreview(func(workDir, stageID string) StagePreviewResult {
		panic("boom")
	}, "/tmp", "add-changes")
	if got.Preview != "" {
		t.Fatalf("panic should soft-fail to empty, got %+v", got)
	}
}

func TestPreviewLogsBecomeNormalLogLines(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	_, _ = m.Update(dashPreviewMsg{
		StageID: "push",
		Text:    "",
		Logs:    []string{"fatal: no upstream configured for branch 'x'"},
	})
	found := false
	for _, ll := range m.logs {
		if strings.Contains(ll.Text, "no upstream") && ll.Stage == "push" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fatal in log ring: %#v", m.logs)
	}
	out := m.renderView()
	if !strings.Contains(out, "no upstream") {
		t.Fatalf("expected Log section to show fatal as normal log:\n%s", out)
	}
}

func TestPreviewCmdsNilWhenNoInjector(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	if cmd := m.previewCmds(); cmd != nil {
		t.Fatalf("expected nil previewCmds without StagePreview, got %T", cmd)
	}
	for _, id := range dashStageIDs {
		if !m.previewSettledFor(id) {
			t.Fatalf("stage %s should be settled when no StagePreview", id)
		}
	}
}

func TestStageInlineMetaPriority(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.applyPreview("add-changes", "3 files")
	m.applyStageEvent(StageEvent{StageID: "add-changes", Kind: "ok", Result: "staged"})
	meta, green := m.stageInlineMeta("add-changes")
	if meta != "result: staged" || !green {
		t.Fatalf("want result: staged green, got %q green=%v", meta, green)
	}
	if strings.Contains(meta, "3 files") {
		t.Fatalf("result should replace preview: %q", meta)
	}

	m.genSubject = "feat: x"
	m.applyPreview("gen-commit-msg", "3 files")
	m.stageResults["gen-commit-msg"] = "feat: x"
	meta, green = m.stageInlineMeta("gen-commit-msg")
	if meta != "msg: feat: x" || !green {
		t.Fatalf("want msg subject, got %q green=%v", meta, green)
	}

	// Settled preview uses prefix.
	m2 := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m2.applyPreview("push", "up to date")
	meta, green = m2.stageInlineMeta("push")
	if meta != "preview: up to date" || green {
		t.Fatalf("want preview: up to date gray, got %q green=%v", meta, green)
	}
}

func TestTuiLogFuncDoesNotUseFmt(t *testing.T) {
	// LogFunc posts to channel only (TUI path), never panics on nil ch.
	tuiLogFunc(nil)("sync", "hello")
	ch := make(chan dashLogLineMsg, 2)
	logf := tuiLogFunc(ch)
	logf("sync", "fetching")
	logf("sync", "") // empty dropped
	msg := <-ch
	if msg.Stage != "sync" || msg.Text != "fetching" {
		t.Fatalf("got %+v", msg)
	}
	select {
	case <-ch:
		t.Fatal("empty line should not be sent")
	default:
	}
}
