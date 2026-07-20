package tui

import (
	"reflect"
	"strings"
	"testing"
)

func TestPlanRecipeStagesOrder(t *testing.T) {
	r := Recipe{
		AddAll:         true,
		GenCommitMsg:   true,
		Commit:         true,
		MergeBack:      true,
		Sync:           true,
		TagNext:        true,
		Push:           true,
		ReinstallLocal: true,
	}
	got := PlanRecipeStages(r)
	want := []string{
		"add-changes", "gen-commit-msg", "commit", "merge-back",
		"sync", "tag-next", "push", "reinstall-local",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan=%v want %v", got, want)
	}
}

func TestPlanRecipeStagesDoneNotBoth(t *testing.T) {
	// Recipe can only have one primary; planner lists whichever flags are set.
	r := Recipe{Done: true, Sync: true}
	got := PlanRecipeStages(r)
	want := []string{"done", "sync"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan=%v want %v", got, want)
	}
}

func TestApplyStageEventUpdatesStageRun(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.applyStageEvent(StageEvent{StageID: "sync", Kind: "start"})
	if m.stageRunState("sync") != StageRunning {
		t.Fatalf("after start: %v", m.stageRunState("sync"))
	}
	if m.status != "running  sync…" {
		t.Fatalf("status=%q", m.status)
	}
	m.applyStageEvent(StageEvent{StageID: "sync", Kind: "ok"})
	if m.stageRunState("sync") != StageOK {
		t.Fatalf("after ok: %v", m.stageRunState("sync"))
	}
	m.applyStageEvent(StageEvent{StageID: "push", Kind: "error", Err: "boom"})
	if m.stageRunState("push") != StageError {
		t.Fatalf("after error: %v", m.stageRunState("push"))
	}
	m.applyStageEvent(StageEvent{StageID: "reinstall-local", Kind: "skipped"})
	if m.stageRunState("reinstall-local") != StageSkipped {
		t.Fatalf("after skipped: %v", m.stageRunState("reinstall-local"))
	}
}

func TestRunAllInjectedPipelineStageEvents(t *testing.T) {
	// Inject RunPipeline that only emits events (no real git/compose).
	opts := RunDashboardOpts{
		WorkDir: "/tmp",
		RunPipeline: func(workDir string, r Recipe, emit func(StageEvent), log LogFunc) error {
			if log != nil {
				log("", "pipeline: sync → push")
			}
			emit(StageEvent{StageID: "sync", Kind: "start"})
			if log != nil {
				log("sync", "start")
			}
			emit(StageEvent{StageID: "sync", Kind: "ok"})
			if log != nil {
				log("sync", "ok")
			}
			emit(StageEvent{StageID: "push", Kind: "start"})
			if log != nil {
				log("push", "start")
			}
			emit(StageEvent{StageID: "push", Kind: "ok"})
			if log != nil {
				log("push", "ok")
			}
			return nil
		},
		ComposeArgv: func(r Recipe) []string { return []string{"--sync", "--push"} },
	}
	m := newTeaDashModel(opts)
	// Minimal recipe: only sync + push so plan matches emits.
	m.sync = true
	m.push = true
	m.tagNext = false
	m.reinstallLocal = false
	m.genCommitMsg = false
	m.commit = false
	m.addAll = false
	m.mainDisabled = true // no merge-back/done

	// Drive events as Update would during a run (without full tea.Program).
	m.resetStageRunForPlan(PlanRecipeStages(m.toRecipe()))
	if m.stageRunState("sync") != StageQueued || m.stageRunState("push") != StageQueued {
		t.Fatalf("queued want sync/push; got sync=%v push=%v", m.stageRunState("sync"), m.stageRunState("push"))
	}

	// Simulate pipeline emit path through applyStageEvent / dashStageEventMsg.
	next, _ := m.Update(dashStageEventMsg{StageID: "sync", Kind: "start"})
	mm := next.(*teaDashModel)
	next, _ = mm.Update(dashStageEventMsg{StageID: "sync", Kind: "ok"})
	mm = next.(*teaDashModel)
	next, _ = mm.Update(dashStageEventMsg{StageID: "push", Kind: "start"})
	mm = next.(*teaDashModel)
	next, _ = mm.Update(dashStageEventMsg{StageID: "push", Kind: "ok"})
	mm = next.(*teaDashModel)

	if mm.stageRunState("sync") != StageOK {
		t.Fatalf("sync state=%v want ok", mm.stageRunState("sync"))
	}
	if mm.stageRunState("push") != StageOK {
		t.Fatalf("push state=%v want ok", mm.stageRunState("push"))
	}

	// Also exercise runDashboardOpCaptured pipeline path end-to-end (stdio capture).
	logCh := make(chan dashLogLineMsg, 8)
	eventCh := make(chan dashStageEventMsg, 16)
	status, err := runDashboardOpCaptured(opts, "/tmp", false, Recipe{Sync: true, Push: true}, "run-all", true, logCh, eventCh)
	if err != nil {
		t.Fatalf("pipeline capture: %v", err)
	}
	if status == "" {
		t.Fatal("expected status")
	}
	var kinds []string
drain:
	for {
		select {
		case ev := <-eventCh:
			kinds = append(kinds, ev.StageID+":"+ev.Kind)
		default:
			break drain
		}
	}
	wantKinds := []string{"sync:start", "sync:ok", "push:start", "push:ok"}
	if !reflect.DeepEqual(kinds, wantKinds) {
		t.Fatalf("events=%v want %v", kinds, wantKinds)
	}
}

func TestResetStageRunForPlan(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.resetStageRunForPlan([]string{"sync", "push"})
	if m.stageRunState("sync") != StageQueued {
		t.Fatalf("sync=%v", m.stageRunState("sync"))
	}
	if m.stageRunState("add-changes") != StageIdle {
		t.Fatalf("add-changes=%v want idle", m.stageRunState("add-changes"))
	}
}

func TestApplyStageEventStoresSubjectAndResult(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.applyStageEvent(StageEvent{
		StageID: "gen-commit-msg",
		Kind:    "ok",
		Subject: "fix: wire StageEvent payload",
		Body:    "Store subject on ok events.\nSecond body line.",
		Result:  "fix: wire StageEvent payload",
	})
	if m.genSubject != "fix: wire StageEvent payload" {
		t.Fatalf("genSubject=%q", m.genSubject)
	}
	if m.genBody != "Store subject on ok events.\nSecond body line." {
		t.Fatalf("genBody=%q", m.genBody)
	}
	if m.stageResults["gen-commit-msg"] != "fix: wire StageEvent payload" {
		t.Fatalf("stageResults gen-commit-msg=%q", m.stageResults["gen-commit-msg"])
	}
	// Body lines should land in the log ring.
	if len(m.logs) < 2 {
		t.Fatalf("expected body lines in logs, got %d: %#v", len(m.logs), m.logs)
	}
	found := false
	for _, ll := range m.logs {
		if strings.Contains(ll.Text, "Store subject") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected body in logs: %#v", m.logs)
	}

	m.applyStageEvent(StageEvent{StageID: "add-changes", Kind: "ok", Result: "staged"})
	if m.stageResults["add-changes"] != "staged" {
		t.Fatalf("stageResults add-changes=%q", m.stageResults["add-changes"])
	}
}

func TestApplyStageEventSubjectRendersInlineMsg(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	m.applyStageEvent(StageEvent{
		StageID: "gen-commit-msg",
		Kind:    "ok",
		Subject: "feat: structured gen-commit result",
	})
	out := m.renderView()
	if !strings.Contains(out, "msg: feat: structured gen-commit result") {
		t.Fatalf("expected msg: subject inline:\n%s", out)
	}
	// msg on same line as gen-commit-msg (compact, no secondary line).
	found := false
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "gen-commit-msg") && strings.Contains(ln, "msg: feat: structured gen-commit result") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("msg should share the gen-commit-msg row:\n%s", out)
	}
}

func TestApplyStageEventResultRendersInline(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.width = 100
	m.color = false
	// Preview would show; result after ok should replace it.
	m.applyPreview("add-changes", "3 files")
	m.applyStageEvent(StageEvent{StageID: "add-changes", Kind: "ok", Result: "staged"})
	out := m.renderView()
	if !strings.Contains(out, "result: staged") {
		t.Fatalf("expected result: staged:\n%s", out)
	}
	// Prefer result over bare preview for that session.
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "add changes") {
			if strings.Contains(ln, "3 files") && !strings.Contains(ln, "result: staged") {
				t.Fatalf("result should replace preview on add-changes:\n%s", ln)
			}
			if !strings.Contains(ln, "result: staged") {
				t.Fatalf("result should be inline on add-changes:\n%s", ln)
			}
		}
	}
}

func TestSplitCommitMessage(t *testing.T) {
	subj, body := SplitCommitMessage("fix: first line\n\nBody paragraph.\nMore.")
	if subj != "fix: first line" {
		t.Fatalf("subject=%q", subj)
	}
	if body != "Body paragraph.\nMore." {
		t.Fatalf("body=%q", body)
	}
	subj, body = SplitCommitMessage("only-subject")
	if subj != "only-subject" || body != "" {
		t.Fatalf("single line: subj=%q body=%q", subj, body)
	}
	subj, body = SplitCommitMessage("")
	if subj != "" || body != "" {
		t.Fatalf("empty: subj=%q body=%q", subj, body)
	}
}

func TestResetStageRunForPlanClearsGenSubject(t *testing.T) {
	m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
	m.genSubject = "old subject"
	m.stageResults["gen-commit-msg"] = "old subject"
	m.resetStageRunForPlan([]string{"gen-commit-msg", "commit"})
	if m.genSubject != "" {
		t.Fatalf("genSubject should clear on re-run plan, got %q", m.genSubject)
	}
	if _, ok := m.stageResults["gen-commit-msg"]; ok {
		t.Fatalf("stageResults should clear gen-commit-msg")
	}
}
