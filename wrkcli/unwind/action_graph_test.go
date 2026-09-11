package unwind

import (
	"strings"
	"testing"
)

func TestBuildActionGraphFiltersLatestDrift(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/leaf", MainRepo: "/tmp/leaf-main", Label: "leaf", Dirty: true, Linked: true},
			{Path: "/tmp/root", MainRepo: "/tmp/root-main", Label: "root", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"leaf", "root"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/leaf", RepoLabel: "leaf", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/root", RepoLabel: "root", LatestTag: "v1.0.0"},
			{Path: "example.com/other", RepoLabel: "other", LatestTag: "v9.0.0"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/root", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/root", To: "example.com/other", Kind: "require", Version: "v1.0.0"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/other", TagOrVersion: "v9.0.0"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true}

	g := BuildActionGraph(snap, flags, ActionGraphOpts{Cleanup: false})
	if g.Excluded < 1 {
		t.Fatalf("expected excluded latest-drift pins, excluded=%d", g.Excluded)
	}
	var modes []string
	var pinReasons []string
	hasTag := false
	hasPropagate := false
	for _, a := range g.Actions {
		modes = append(modes, string(a.Mode))
		if a.Mode == ModeTagNext {
			hasTag = true
		}
		if pinLike(a.Mode) {
			pinReasons = append(pinReasons, string(a.Reason))
			if a.Reason == ReasonPropagate {
				hasPropagate = true
			}
			if a.Reason == ReasonLatestDrift {
				t.Fatalf("default filter must omit latest-drift pin: %+v", a)
			}
		}
	}
	if !hasTag {
		t.Fatalf("missing tag-next: %v", modes)
	}
	if !hasPropagate {
		t.Fatalf("missing propagate pin: reasons=%v modes=%v", pinReasons, modes)
	}

	// Edge: tag leaf before pin root←leaf
	tagID := "tag-next:example.com/leaf"
	pinID := "pin:example.com/root<example.com/leaf"
	pin := findAction(g, pinID)
	if pin == nil {
		t.Fatalf("missing %s", pinID)
	}
	if pin.Mode != ModeDepUpdate {
		t.Fatalf("cross-repo pin mode=%s want dep-update", pin.Mode)
	}
	if !depsContain(pin.Deps, tagID) {
		t.Fatalf("pin deps=%v want %s", pin.Deps, tagID)
	}

	g2 := BuildActionGraph(snap, flags, ActionGraphOpts{Cleanup: true})
	if g2.Excluded != 0 {
		t.Fatalf("cleanup should include all pins, excluded=%d", g2.Excluded)
	}
	foundDrift := false
	for _, a := range g2.Actions {
		if pinLike(a.Mode) && a.Reason == ReasonLatestDrift {
			foundDrift = true
		}
	}
	if !foundDrift {
		t.Fatal("cleanup graph should include latest-drift pin")
	}
}

func TestBuildActionGraphLandChain(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/a", MainRepo: "/tmp/a-main", Label: "a", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"a"}, NeedsLand: true},
	}
	g := BuildActionGraph(snap, UnwindFlags{MergeBack: true, GenCommitMsg: true}, ActionGraphOpts{})
	gen := findAction(g, "gen-commit:a")
	mb := findAction(g, "merge-back:a")
	if gen == nil || mb == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if !depsContain(mb.Deps, gen.ID) {
		t.Fatalf("merge-back deps=%v want gen-commit", mb.Deps)
	}
	// No peel banner action.
	for _, a := range g.Actions {
		if a.Mode == "peel" {
			t.Fatal("peel must not be an action mode")
		}
	}
}

func TestBuildActionGraphDeferredWaitsOnCascadePins(t *testing.T) {
	t.Parallel()
	// leaf = free host (early); root = pure pin-consumer (deferred).
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/leaf", MainRepo: "/tmp/leaf-main", Label: "leaf", Dirty: true, Linked: true},
			{Path: "/tmp/root", MainRepo: "/tmp/root-main", Label: "root", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"leaf", "root"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/leaf", RepoLabel: "leaf", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/root", RepoLabel: "root", LatestTag: "v1.0.0", OwnedChanged: true, NextTag: "v1.0.1"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/root", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/root", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})

	genRoot := findAction(g, "gen-commit:root")
	pin := findAction(g, "pin:example.com/root<example.com/leaf")
	mbLeaf := findAction(g, "merge-back:leaf")
	if genRoot == nil || pin == nil || mbLeaf == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	tagLeaf := findAction(g, "tag-next:example.com/leaf")
	if tagLeaf == nil {
		t.Fatalf("missing tag-next leaf, actions=%v", actionIDs(g))
	}
	// B1: deferred gen-commit waits on cascade pin into root (not the reverse).
	if pin.Mode != ModeDepUpdate {
		t.Fatalf("deferred cross-repo pin mode=%s want dep-update", pin.Mode)
	}
	if !depsContain(genRoot.Deps, pin.ID) {
		t.Fatalf("deferred gen-commit deps=%v want pin %s", genRoot.Deps, pin.ID)
	}
	if depsContain(pin.Deps, "merge-back:root") || depsContain(pin.Deps, "gen-commit:root") {
		t.Fatalf("pin must not wait on deferred land: deps=%v", pin.Deps)
	}
	// Pin gate is tag-next (not redundant merge-back); tag already waits on land.
	if !depsContain(pin.Deps, tagLeaf.ID) {
		t.Fatalf("deferred pin deps=%v want %s", pin.Deps, tagLeaf.ID)
	}
	if depsContain(pin.Deps, mbLeaf.ID) {
		t.Fatalf("pin must not also dep merge-back when tag-next covers it: deps=%v", pin.Deps)
	}
	if depsContain(genRoot.Deps, mbLeaf.ID) {
		t.Fatalf("deferred gen-commit must not depend on early merge-back: %v", genRoot.Deps)
	}
	tagRoot := findAction(g, "tag-next:example.com/root")
	if tagRoot == nil {
		t.Fatalf("missing tag-next root, actions=%v", actionIDs(g))
	}
	if depsContain(tagRoot.Deps, pin.ID) {
		t.Fatalf("tag-next(consumer) must not chord pin when land covers it: %v", tagRoot.Deps)
	}
	if !depsContain(tagRoot.Deps, "merge-back:root") {
		t.Fatalf("tag-next root deps=%v want merge-back:root", tagRoot.Deps)
	}
	// Stretch: pin after tag after merge; gen-commit stays prep (0).
	if pin.Rank <= tagLeaf.Rank {
		t.Fatalf("deferred pin rank=%d should be after tag rank=%d", pin.Rank, tagLeaf.Rank)
	}
	if tagLeaf.Rank <= mbLeaf.Rank {
		t.Fatalf("tag rank=%d should be after merge rank=%d", tagLeaf.Rank, mbLeaf.Rank)
	}
	if genRoot.Rank != 0 {
		t.Fatalf("gen-commit band rank=%d want 0 (prep)", genRoot.Rank)
	}
}

func TestBuildActionGraphCrossRepoGateAndBands(t *testing.T) {
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
			{Path: "example.com/app", RepoLabel: "app", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/app", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/app", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true, Sync: true, ReinstallLocal: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	if g.LaneLevels["dep"] != 0 || g.LaneLevels["app"] != 1 {
		t.Fatalf("lane_levels=%v want dep=0 app=1", g.LaneLevels)
	}
	genApp := findAction(g, "gen-commit:app")
	tagDep := findAction(g, "tag-next:example.com/dep")
	if genApp == nil || tagDep == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	pinApp := findAction(g, "pin:example.com/app<example.com/dep")
	if pinApp == nil {
		t.Fatalf("missing dep-update, actions=%v", actionIDs(g))
	}
	if pinApp.Mode != ModeDepUpdate {
		t.Fatalf("cross-repo pin mode=%s want dep-update", pinApp.Mode)
	}
	if !depsContain(pinApp.Deps, tagDep.ID) {
		t.Fatalf("dep-update deps=%v want %s", pinApp.Deps, tagDep.ID)
	}
	if !depsContain(genApp.Deps, pinApp.ID) {
		t.Fatalf("app gen-commit deps=%v want dep-update %s", genApp.Deps, pinApp.ID)
	}
	if depsContain(genApp.Deps, tagDep.ID) {
		t.Fatalf("app gen-commit must not depend on tag-next directly: %v", genApp.Deps)
	}
	if depsContain(genApp.Deps, "merge-back:dep") {
		t.Fatalf("app gen-commit must not depend on dep merge-back: %v", genApp.Deps)
	}
	pushDep := findAction(g, "push:dep")
	pushApp := findAction(g, "push:app")
	if pushDep == nil || pushApp == nil {
		t.Fatalf("want per-project push, actions=%v", actionIDs(g))
	}
	if findAction(g, "push:mains") != nil {
		t.Fatal("global push:mains should be gone")
	}
	ri := findAction(g, "reinstall:dep")
	if ri == nil {
		t.Fatalf("missing reinstall-local, actions=%v", actionIDs(g))
	}
	if depsContain(ri.Deps, "push:dep") || depsContain(ri.Deps, "sync:dep") {
		t.Fatalf("reinstall must not wait on ship: %v", ri.Deps)
	}
	// Band floors + stretch: merge at floor, tag/pin/ship move right along deps.
	mbDep := findAction(g, "merge-back:dep")
	if mbDep.Rank != 1 {
		t.Fatalf("dep merge rank=%d want floor 1", mbDep.Rank)
	}
	if tagDep.Rank <= mbDep.Rank {
		t.Fatalf("dep tag rank=%d should be after merge %d", tagDep.Rank, mbDep.Rank)
	}
	if pushDep.Rank <= tagDep.Rank {
		t.Fatalf("dep push rank=%d should be after tag %d", pushDep.Rank, tagDep.Rank)
	}
	mbApp := findAction(g, "merge-back:app")
	if mbApp == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if mbApp.Rank != 2 {
		t.Fatalf("app merge rank=%d want floor 2", mbApp.Rank)
	}
	if pinApp.Rank <= tagDep.Rank {
		t.Fatalf("app pin rank=%d should be after dep tag %d", pinApp.Rank, tagDep.Rank)
	}
	if depsContain(pinApp.Deps, mbDep.ID) {
		t.Fatalf("app pin must not dep dep merge-back when tag covers it: deps=%v", pinApp.Deps)
	}
	tagApp := findAction(g, "tag-next:example.com/app")
	if tagApp == nil {
		t.Fatalf("missing tag-next app, actions=%v", actionIDs(g))
	}
	if depsContain(tagApp.Deps, pinApp.ID) {
		t.Fatalf("tag-next app must not chord pin when land covers it: %v", tagApp.Deps)
	}
	if pushApp.Rank <= pinApp.Rank {
		t.Fatalf("app push rank=%d should be after pin %d", pushApp.Rank, pinApp.Rank)
	}
}

func TestBuildActionGraphDeferredPinSkipsOwnMergeGate(t *testing.T) {
	t.Parallel()
	// Deferred root waits on pin; pin of same-lane untagged dep must not wait on
	// merge-back:root (that closes gen→merge→pin→gen).
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/leaf", MainRepo: "/tmp/leaf-main", Label: "leaf", Dirty: true, Linked: true},
			{Path: "/tmp/root", MainRepo: "/tmp/root-main", Label: "root", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"leaf", "root"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/leaf", RepoLabel: "leaf", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/root", RepoLabel: "root", LatestTag: "v1.0.0", OwnedChanged: true, NextTag: "v1.0.1"},
			{Path: "example.com/root/sidecar", RepoLabel: "root", LatestTag: "v0.1.0"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/root", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/root", To: "example.com/root/sidecar", Kind: "require", Version: "v0.1.0"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/root/sidecar", TagOrVersion: "v0.1.0"},
			{Kind: CascadeTagNext, ModulePath: "example.com/root", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	pinSide := findAction(g, "pin:example.com/root<example.com/root/sidecar")
	mbRoot := findAction(g, "merge-back:root")
	genRoot := findAction(g, "gen-commit:root")
	if pinSide == nil || mbRoot == nil || genRoot == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if depsContain(pinSide.Deps, mbRoot.ID) {
		t.Fatalf("same-lane untagged pin must not dep own merge: deps=%v", pinSide.Deps)
	}
	if depsContain(pinSide.Deps, "merge-back:leaf") {
		t.Fatalf("pin-before-tag intra must not dep early merge-back: deps=%v", pinSide.Deps)
	}
	if !depsContain(genRoot.Deps, pinSide.ID) {
		t.Fatalf("deferred gen should wait on pin: deps=%v", genRoot.Deps)
	}
	tagRoot := findAction(g, "tag-next:example.com/root")
	if tagRoot == nil {
		t.Fatalf("missing tag-next root, actions=%v", actionIDs(g))
	}
	if depsContain(tagRoot.Deps, pinSide.ID) {
		t.Fatalf("tag-next must not chord intra pin-before-tag when land covers it: %v", tagRoot.Deps)
	}
	// Stretch must terminate (no infinite loop on residual cycles).
	if pinSide.Rank < 0 || mbRoot.Rank < 0 {
		t.Fatalf("ranks unset pin=%d merge=%d", pinSide.Rank, mbRoot.Rank)
	}
}

func TestBuildActionGraphSameLanePinAfterTag(t *testing.T) {
	t.Parallel()
	// go-pkgs tags; go-pkgs/cmd pins go-pkgs in the same repo lane.
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/dot", MainRepo: "/tmp/dot-main", Label: "dot-pkgs", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"dot-pkgs"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/go-pkgs", RepoLabel: "dot-pkgs", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/go-pkgs/cmd", RepoLabel: "dot-pkgs", LatestTag: "v0.0.1"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/go-pkgs/cmd", To: "example.com/go-pkgs", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/go-pkgs", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/go-pkgs/cmd", DepModulePath: "example.com/go-pkgs", TagOrVersion: "v0.0.2"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	tag := findAction(g, "tag-next:example.com/go-pkgs")
	pin := findAction(g, "pin:example.com/go-pkgs/cmd<example.com/go-pkgs")
	mb := findAction(g, "merge-back:dot-pkgs")
	if tag == nil || pin == nil || mb == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if pin.Mode != ModePin {
		t.Fatalf("intra pin mode=%s want pin", pin.Mode)
	}
	if !depsContain(pin.Deps, tag.ID) {
		t.Fatalf("pin deps=%v want %s", pin.Deps, tag.ID)
	}
	if depsContain(pin.Deps, mb.ID) {
		t.Fatalf("same-lane pin must not also dep merge-back: deps=%v", pin.Deps)
	}
	if pin.Rank <= tag.Rank {
		t.Fatalf("pin rank=%d want after tag rank=%d", pin.Rank, tag.Rank)
	}
}

func TestBuildActionGraphEmitsDepUpdateWhenCascadePinMissing(t *testing.T) {
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
			{Path: "example.com/app", RepoLabel: "app", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/app", To: "example.com/dep", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/dep", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/app", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	pin := findAction(g, "pin:example.com/app<example.com/dep")
	tag := findAction(g, "tag-next:example.com/dep")
	gen := findAction(g, "gen-commit:app")
	if pin == nil || tag == nil || gen == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if pin.Mode != ModeDepUpdate {
		t.Fatalf("mode=%s want dep-update", pin.Mode)
	}
	if !depsContain(pin.Deps, tag.ID) {
		t.Fatalf("dep-update deps=%v want %s", pin.Deps, tag.ID)
	}
	if !depsContain(gen.Deps, pin.ID) {
		t.Fatalf("gen-commit deps=%v want %s", gen.Deps, pin.ID)
	}
	if depsContain(gen.Deps, tag.ID) {
		t.Fatalf("gen-commit must not depend on tag-next directly: %v", gen.Deps)
	}
}

func TestBuildActionGraphPinBeforeTagLandCovers(t *testing.T) {
	t.Parallel()
	// 8082 shape: tagged leaf, tagged root, untagged sdk catch-up pin.
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/leaf", MainRepo: "/tmp/leaf-main", Label: "leaf", Dirty: true, Linked: true},
			{Path: "/tmp/root", MainRepo: "/tmp/root-main", Label: "root", Dirty: true, Linked: true},
			{Path: "/tmp/sdk", MainRepo: "/tmp/sdk-main", Label: "sdk", Dirty: false, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"leaf", "root"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/leaf", RepoLabel: "leaf", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/root", RepoLabel: "root", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
			{Path: "example.com/sdk", RepoLabel: "sdk", LatestTag: "v1.0.36"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/root", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/root", To: "example.com/sdk", Kind: "require", Version: "v1.0.30"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/sdk", TagOrVersion: "v1.0.36"},
			{Kind: CascadeTagNext, ModulePath: "example.com/root", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	pinLeaf := findAction(g, "pin:example.com/root<example.com/leaf")
	pinSDK := findAction(g, "pin:example.com/root<example.com/sdk")
	tagRoot := findAction(g, "tag-next:example.com/root")
	genRoot := findAction(g, "gen-commit:root")
	if pinLeaf == nil || tagRoot == nil || genRoot == nil {
		t.Fatalf("actions=%v", actionIDs(g))
	}
	if pinSDK != nil {
		t.Fatalf("sdk catch-up pin must be omitted without cleanup, got reason=%s", pinSDK.Reason)
	}
	if g.Excluded < 1 {
		t.Fatalf("want excluded latest-drift sdk pin, excluded=%d", g.Excluded)
	}
	if depsContain(tagRoot.Deps, pinLeaf.ID) {
		t.Fatalf("tag-next root must not chord pins: %v", tagRoot.Deps)
	}
	if !depsContain(tagRoot.Deps, "merge-back:root") {
		t.Fatalf("tag-next root deps=%v want merge-back:root", tagRoot.Deps)
	}
	if !depsContain(genRoot.Deps, pinLeaf.ID) {
		t.Fatalf("gen-commit should wait on propagate pin: %v", genRoot.Deps)
	}

	g2 := BuildActionGraph(snap, flags, ActionGraphOpts{Cleanup: true})
	pinSDK = findAction(g2, "pin:example.com/root<example.com/sdk")
	if pinSDK == nil {
		t.Fatalf("cleanup should include sdk catch-up, actions=%v", actionIDs(g2))
	}
	if pinSDK.Reason != ReasonLatestDrift {
		t.Fatalf("sdk pin reason=%s want latest-drift", pinSDK.Reason)
	}
	if pinSDK.Mode != ModeDepUpdate {
		t.Fatalf("sdk pin mode=%s want dep-update", pinSDK.Mode)
	}
}

func TestBuildActionGraphIntraRequireDriftIsPropagate(t *testing.T) {
	t.Parallel()
	// Both directions of same-repo LatestTag catch-up must pin without --cleanup:
	// nested←ancestor (agent-pro cmd←parent) and parent←nested (CS-openterm2).
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/free", MainRepo: "/tmp/free-main", Label: "free", Dirty: true, Linked: true},
			{Path: "/tmp/app", MainRepo: "/tmp/app-main", Label: "app", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"free", "app"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/dot-pkgs", RepoLabel: "free", LatestTag: "v0.0.2"},
			{Path: "example.com/dot-pkgs/cmd-harness", RepoLabel: "free", LatestTag: "v0.0.1"},
			{Path: "example.com/app", RepoLabel: "app", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
			{Path: "example.com/app/pkgs/log", RepoLabel: "app", LatestTag: "v0.0.2"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/dot-pkgs/cmd-harness", To: "example.com/dot-pkgs", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/app", To: "example.com/dot-pkgs", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/app", To: "example.com/app/pkgs/log", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadePin, ModulePath: "example.com/dot-pkgs/cmd-harness", DepModulePath: "example.com/dot-pkgs", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/dot-pkgs", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/app", DepModulePath: "example.com/app/pkgs/log", TagOrVersion: "v0.0.2"},
			{Kind: CascadeTagNext, ModulePath: "example.com/app", TagOrVersion: "v1.0.1"},
		}},
	}
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true}
	g := BuildActionGraph(snap, flags, ActionGraphOpts{})
	cmdPin := findAction(g, "pin:example.com/dot-pkgs/cmd-harness<example.com/dot-pkgs")
	logPin := findAction(g, "pin:example.com/app<example.com/app/pkgs/log")
	if cmdPin == nil || logPin == nil {
		t.Fatalf("intra catch-up pins must be included without cleanup, actions=%v", actionIDs(g))
	}
	if cmdPin.Reason != ReasonPropagate || logPin.Reason != ReasonPropagate {
		t.Fatalf("cmd reason=%s log reason=%s want propagate", cmdPin.Reason, logPin.Reason)
	}
	// Cross-repo catch-up to already-released free stays latest-drift (omitted).
	if findAction(g, "pin:example.com/app<example.com/dot-pkgs") != nil {
		t.Fatal("cross-repo LatestTag catch-up must be omitted without cleanup")
	}
}

func TestBuildActionGraphPushForTagOnlyAlreadyMain(t *testing.T) {
	t.Parallel()
	// Already on main: dirty + NextTag, no linked land — ship must still push tags.
	snap := &Snapshot{
		WorkDir: "/tmp/root",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/root", MainRepo: "/tmp/root", Label: "root", Dirty: true, Linked: false},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"root"}, NeedsLand: false},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/root", RepoLabel: "root", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/root", TagOrVersion: "v0.0.2"},
		}},
	}
	g := BuildActionGraph(snap, UnwindFlags{TagNext: true, Push: true}, ActionGraphOpts{})
	if findAction(g, "tag-next:example.com/root") == nil {
		t.Fatalf("missing tag-next, actions=%v", actionIDs(g))
	}
	push := findAction(g, "push:root")
	if push == nil {
		t.Fatalf("tag-only already-main must emit push, actions=%v", actionIDs(g))
	}
	if !depsContain(push.Deps, "tag-next:example.com/root") {
		t.Fatalf("push deps=%v want tag-next", push.Deps)
	}
}

func TestJobPlanUsesActionGraphEpochs(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/a", Label: "a", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"a"}, NeedsLand: true},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/a", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/b", DepModulePath: "example.com/c", TagOrVersion: "v1.0.0"},
		}},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/a", RepoLabel: "a", NextTag: "v0.0.2", OwnedChanged: true},
			{Path: "example.com/b", RepoLabel: "b", LatestTag: "v0.0.1"},
			{Path: "example.com/c", RepoLabel: "c", LatestTag: "v1.0.0"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/b", To: "example.com/c", Kind: "require", Version: "v0.0.9"},
		},
	}
	job := BuildJobPlan(snap, UnwindFlags{MergeBack: true, TagNext: true})
	if job.ActionGraph == nil || len(job.ActionGraph.Actions) == 0 {
		t.Fatal("expected action_graph")
	}
	if job.ActionGraph.Excluded < 1 {
		t.Fatalf("expected excluded drift pin, excluded=%d", job.ActionGraph.Excluded)
	}
	raw := FormatActionGraphDebug(job.ActionGraph)
	if !strings.Contains(raw, "tag-next") {
		t.Fatalf("debug missing tag-next:\n%s", raw)
	}
}

func findAction(g *ActionGraph, id string) *Action {
	if g == nil {
		return nil
	}
	for _, a := range g.Actions {
		if a.ID == id {
			return a
		}
	}
	return nil
}

func depsContain(deps []string, id string) bool {
	for _, d := range deps {
		if d == id {
			return true
		}
	}
	return false
}

func actionIDs(g *ActionGraph) []string {
	var out []string
	for _, a := range g.Actions {
		out = append(out, a.ID)
	}
	return out
}
