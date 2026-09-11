package unwind

import "testing"

func TestBuildJobPhasesSplitsCrossAndIntraPins(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/dep", MainRepo: "/tmp/dep-main", Label: "dep", Dirty: true, Linked: true},
			{Path: "/tmp/app", MainRepo: "/tmp/app-main", Label: "app", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"dep", "app"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/dep", RepoLabel: "dep", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/dep/cmd", RepoLabel: "dep", LatestTag: "v0.0.1"},
			{Path: "example.com/app", RepoLabel: "app", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/app", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/dep/cmd", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/dep/cmd", DepModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/app", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true, Sync: true, ReinstallLocal: true}
	phases := buildJobPhases(snap, flags, ActionGraphOpts{})
	if len(phases) != 4 {
		t.Fatalf("phases=%d want 4", len(phases))
	}
	if got := phaseByID(phases, "snapshot"); got == nil {
		t.Fatal("missing snapshot phase")
	}
	reposPh := phaseByID(phases, "repos")
	modsPh := phaseByID(phases, "modules")
	shipPh := phaseByID(phases, "ship")
	if reposPh == nil || reposPh.ActionGraph == nil {
		t.Fatalf("repos=%+v", reposPh)
	}
	if modsPh == nil || modsPh.ByRepo == nil {
		t.Fatalf("modules=%+v", modsPh)
	}
	if shipPh == nil || shipPh.ByRepo == nil {
		t.Fatalf("ship=%+v", shipPh)
	}
	if reposPh.Title != phaseTitleRepos || modsPh.Title != phaseTitleModules || shipPh.Title != phaseTitleShip {
		t.Fatalf("titles repos=%q modules=%q ship=%q", reposPh.Title, modsPh.Title, shipPh.Title)
	}
	p1 := reposPh.ActionGraph
	if findAction(p1, "push:dep") != nil || findAction(p1, "sync:dep") != nil {
		t.Fatalf("phase1 must not include push/sync, actions=%v", actionIDs(p1))
	}
	if findAction(p1, "reinstall:dep") == nil {
		t.Fatalf("phase1 missing reinstall-local, actions=%v", actionIDs(p1))
	}
	if findAction(shipPh.ByRepo["dep"], "push:dep") == nil || findAction(shipPh.ByRepo["dep"], "sync:dep") == nil {
		t.Fatalf("ship dep missing push/sync, actions=%v", actionIDs(shipPh.ByRepo["dep"]))
	}
	if findAction(shipPh.ByRepo["dep"], "reinstall:dep") != nil {
		t.Fatal("ship must not include reinstall-local")
	}
	if findAction(p1, "pin:example.com/dep/cmd<example.com/dep") != nil {
		t.Fatal("phase1 must not include intra pin dep/cmd<-dep")
	}
	cross := findAction(p1, "pin:example.com/app<example.com/dep")
	if cross == nil {
		t.Fatalf("phase1 missing cross dep-update, actions=%v", actionIDs(p1))
	}
	if cross.Mode != ModeDepUpdate {
		t.Fatalf("phase1 cross mode=%s want dep-update", cross.Mode)
	}
	if findAction(p1, "gen-commit:dep#gen-commit-msg") == nil || findAction(p1, "tag-next:example.com/dep") == nil {
		t.Fatalf("phase1 missing land/tag, actions=%v", actionIDs(p1))
	}
	depMods := modsPh.ByRepo["dep"]
	if depMods == nil {
		t.Fatalf("phase2 by_repo keys=%v", keysOf(modsPh.ByRepo))
	}
	intra := findAction(depMods, "pin:example.com/dep/cmd<example.com/dep")
	if intra == nil {
		t.Fatalf("phase2 dep missing intra pin, actions=%v", actionIDs(depMods))
	}
	if intra.Mode != ModeDepUpdate {
		t.Fatalf("phase2 intra mode=%s want dep-update", intra.Mode)
	}
	cmt := findAction(depMods, "commit:dep")
	if cmt == nil {
		t.Fatalf("phase2 missing commit card, actions=%v", actionIDs(depMods))
	}
	if cmt.Mode != ModeCommit {
		t.Fatalf("commit mode=%s want commit", cmt.Mode)
	}
	wantMsg := "dep: example.com/dep v0.0.1 -> v0.0.2"
	if cmt.Detail != wantMsg {
		t.Fatalf("commit detail=%q want %q", cmt.Detail, wantMsg)
	}
	if !depsContain(cmt.Deps, intra.ID) {
		t.Fatalf("commit deps=%v want %s", cmt.Deps, intra.ID)
	}
	appP2 := modsPh.ByRepo["app"]
	if appP2 == nil {
		t.Fatalf("phase2 should list app even when empty, keys=%v", keysOf(modsPh.ByRepo))
	}
	if len(appP2.Actions) != 0 {
		t.Fatalf("app phase2 want empty, actions=%v", actionIDs(appP2))
	}
	if appP2.FilterNote != phase2EmptyNote {
		t.Fatalf("app filter_note=%q want %q", appP2.FilterNote, phase2EmptyNote)
	}

	// Intra require-drift catch-up (parent←nested LatestTag, CS-openterm2) is
	// propagate: included without cleanup, Phase 2 (intra), never Phase 1.
	snap.ModuleNodes = append(snap.ModuleNodes, UnwindGraphModuleNode{
		Path: "example.com/app/pkgs/log", RepoLabel: "app", LatestTag: "v0.0.2",
	})
	snap.ModuleEdges = append(snap.ModuleEdges, UnwindGraphModuleEdge{
		From: "example.com/app", To: "example.com/app/pkgs/log", Kind: "require", Version: "v0.0.1",
	})
	snap.Cascade.Steps = append(snap.Cascade.Steps, UnwindCascadeStep{
		Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/app/pkgs/log", TagOrVersion: "v0.0.2",
	})
	logPinID := "pin:example.com/app<example.com/app/pkgs/log"
	phases = buildJobPhases(snap, flags, ActionGraphOpts{})
	p1 = phaseByID(phases, "repos").ActionGraph
	if findAction(p1, logPinID) != nil {
		t.Fatalf("intra catch-up pin must be Phase 2, not phase1, actions=%v", actionIDs(p1))
	}
	appMods := phaseByID(phases, "modules").ByRepo["app"]
	catchUp := findAction(appMods, logPinID)
	if catchUp == nil {
		t.Fatalf("phase2 missing intra catch-up without cleanup, actions=%v", actionIDs(appMods))
	}
	if catchUp.Mode != ModeDepUpdate {
		t.Fatalf("phase2 intra catch-up mode=%s want dep-update", catchUp.Mode)
	}
	if catchUp.Reason != ReasonPropagate {
		t.Fatalf("reason=%s want propagate", catchUp.Reason)
	}
	phases = buildJobPhases(snap, flags, ActionGraphOpts{Cleanup: true})
	p1 = phaseByID(phases, "repos").ActionGraph
	if findAction(p1, logPinID) != nil {
		t.Fatalf("cleanup: intra catch-up stays Phase 2, not phase1, actions=%v", actionIDs(p1))
	}
	if findAction(phaseByID(phases, "modules").ByRepo["app"], logPinID) == nil {
		t.Fatal("cleanup: phase2 still has intra catch-up")
	}

	job := BuildJobPlan(snap, flags)
	if len(job.Phases) != 4 {
		t.Fatalf("JobPlan.phases=%d", len(job.Phases))
	}
	if job.ActionGraph == nil || findAction(job.ActionGraph, "pin:example.com/dep/cmd<example.com/dep") != nil {
		t.Fatal("JobPlan.action_graph should be phase1 (no intra pin)")
	}
	if findAction(job.ActionGraph, "push:dep") != nil {
		t.Fatal("JobPlan.action_graph should not include ship push")
	}
}

func keysOf(m map[string]*ActionGraph) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
