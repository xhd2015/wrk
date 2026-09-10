package unwind

import (
	"testing"
)

func TestExpandLandActionsAddAllCommit(t *testing.T) {
	raw := []*Action{
		{ID: "gen-commit:a", Mode: ModeGenCommit, Lane: "a", Reason: ReasonLand,
			Subject: Subject{Display: "a", RepoLabel: "a"}},
		{ID: "merge-back:a", Mode: ModeMergeBack, Lane: "a", Reason: ReasonLand,
			Deps: []string{"gen-commit:a"}, Subject: Subject{Display: "a", RepoLabel: "a"}},
		{ID: "pin:x", Mode: ModeDepUpdate, Lane: "b", Reason: ReasonPropagate,
			Deps: []string{"tag"}, Subject: Subject{Display: "x"}},
		{ID: "gen-commit:b", Mode: ModeGenCommit, Lane: "b", Reason: ReasonLand,
			Deps: []string{"pin:x"}, Subject: Subject{Display: "b", RepoLabel: "b"}},
	}
	out := ExpandLandActions(raw, true, true)
	byID := map[string]*Action{}
	for _, a := range out {
		byID[a.ID] = a
	}
	if byID["gen-commit:a"] != nil {
		t.Fatal("bundled gen-commit should be removed")
	}
	add := byID["gen-commit:a#add-all"]
	msg := byID["gen-commit:a#gen-commit-msg"]
	cmt := byID["gen-commit:a#commit"]
	if add == nil || add.Mode != ModeAddAll {
		t.Fatalf("missing add-all: %v", add)
	}
	if msg == nil || msg.Mode != ModeGenCommitMsg {
		t.Fatalf("missing gen-commit-msg: %v", msg)
	}
	if !depsContain(msg.Deps, add.ID) {
		t.Fatalf("msg deps=%v want add-all", msg.Deps)
	}
	msgB := byID["gen-commit:b#gen-commit-msg"]
	if msgB == nil || !depsContain(msgB.Deps, "pin:x") || !depsContain(msgB.Deps, "gen-commit:b#add-all") {
		t.Fatalf("b msg deps=%v want pin + add-all", msgB.Deps)
	}
	if len(msg.Produces) != 1 || msg.Produces[0] != "message:a" {
		t.Fatalf("msg produces=%v", msg.Produces)
	}
	if cmt == nil || cmt.Mode != ModeCommit {
		t.Fatalf("missing commit: %v", cmt)
	}
	if !depsContain(cmt.Deps, msg.ID) {
		t.Fatalf("commit deps=%v want msg", cmt.Deps)
	}
	mb := byID["merge-back:a"]
	if mb == nil || !depsContain(mb.Deps, cmt.ID) {
		t.Fatalf("merge-back deps=%v want land commit", mb.Deps)
	}
	// Deferred land: commit waits on pin; pin waits on add-all.
	cmtB := byID["gen-commit:b#commit"]
	if cmtB == nil || !depsContain(cmtB.Deps, "pin:x") {
		t.Fatalf("b commit deps=%v want pin", cmtB.Deps)
	}
	pin := byID["pin:x"]
	if pin == nil || !depsContain(pin.Deps, "gen-commit:b#add-all") {
		t.Fatalf("pin deps=%v want b add-all", pin.Deps)
	}
}

func TestResolveJobs(t *testing.T) {
	if resolveJobs(0, 8) != 8 {
		t.Fatal("0 → GOMAXPROCS")
	}
	if resolveJobs(1, 8) != 1 {
		t.Fatal("1 → serial")
	}
	if resolveJobs(4, 8) != 4 {
		t.Fatal("explicit jobs")
	}
	if resolveJobs(-1, 8) != 1 {
		t.Fatal("negative → 1")
	}
}
