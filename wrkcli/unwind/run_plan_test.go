package unwind

import (
	"strings"
	"testing"
)

func sampleJobSnap() *Snapshot {
	return &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/stack/leaf", MainRepo: "/tmp/leaf-main", Label: "leaf", Dirty: true, Linked: true},
			{Path: "/tmp/stack/root", MainRepo: "/tmp/root-main", Label: "root", Dirty: true, Linked: true},
			{Path: "/tmp/stack/sdk", MainRepo: "/tmp/sdk-main", Label: "sdk", Dirty: false, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"leaf", "root"}, NeedsLand: true},
		ModuleNodes: []UnwindGraphModuleNode{
			{Path: "example.com/leaf", RepoLabel: "leaf", NextTag: "v0.0.2", LatestTag: "v0.0.1", OwnedChanged: true},
			{Path: "example.com/leaf/cmd", RepoLabel: "leaf", LatestTag: "v0.0.1"},
			{Path: "example.com/root", RepoLabel: "root", NextTag: "v1.0.1", LatestTag: "v1.0.0", OwnedChanged: true},
			{Path: "example.com/sdk", RepoLabel: "sdk", LatestTag: "v1.0.36"},
		},
		ModuleEdges: []UnwindGraphModuleEdge{
			{From: "example.com/root", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
			{From: "example.com/root", To: "example.com/sdk", Kind: "require", Version: "v1.0.30"},
			{From: "example.com/leaf/cmd", To: "example.com/leaf", Kind: "require", Version: "v0.0.1"},
		},
		Cascade: &UnwindCascadePlan{Steps: []UnwindCascadeStep{
			{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/leaf/cmd", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
			{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/sdk", TagOrVersion: "v1.0.36"},
			{Kind: CascadeTagNext, ModulePath: "example.com/root", TagOrVersion: "v1.0.1"},
		}},
	}
}

func TestJobRunnerDryRunGroupsByRepo(t *testing.T) {
	snap := sampleJobSnap()
	flags := UnwindFlags{
		MergeBack: true, TagNext: true, GenCommitMsg: true,
		Push: true, Sync: true, ReinstallLocal: true, AddAll: true, DryRun: true,
		GenCommitArgs: []string{"--commit"},
	}
	job := BuildJobPlan(snap, flags)
	var outBuf, errBuf strings.Builder
	st := newStageWriter(jobPlanStageTotal, false, false)
	st.outW = &outBuf
	st.errW = &errBuf
	r := &jobRunner{
		workDir: "/tmp/stack",
		snap:    snap,
		job:     job,
		flags:   flags,
		dryRun:  true,
		st:      st,
		byLabel: pickPeelMembersByLabel(snap.Inv.Members),
		peeled:  map[string]bool{},
	}
	if err := r.run(); err != nil {
		t.Fatal(err)
	}
	stderr := errBuf.String()
	stdout := outBuf.String()
	for _, want := range []string{
		"[2/4] phase-1 · cross-repo unwind\n",
		"[3/4] phase-2 · intra-repo update\n",
		"[4/4] ship\n",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q:\n%s", want, stderr)
		}
	}
	if strings.Contains(stderr, "\x1b[") {
		t.Fatalf("stage headers must be uncolored:\n%s", stderr)
	}
	if strings.Contains(stdout, "would: peel") {
		t.Fatalf("peel jargon in stdout:\n%s", stdout)
	}
	if i := strings.Index(stdout, "example.com/leaf/cmd"); i >= 0 {
		// Phase 2 pin subject must sit under the leaf repo header, not a module-path group.
		before := stdout[:i]
		if !strings.Contains(before, "leaf\n") {
			t.Fatalf("phase-2 pin not under leaf repo:\n%s", stdout)
		}
	}
	if !strings.Contains(stdout, "would: git add -A") {
		t.Fatalf("missing git add -A:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: gen-commit-msg") {
		t.Fatalf("missing gen-commit-msg:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: merge-back") {
		t.Fatalf("missing merge-back:\n%s", stdout)
	}
	if strings.Contains(stdout, "would: merge-back ") {
		t.Fatalf("merge-back should not repeat the path:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: dep-update example.com/root <- example.com/leaf") {
		t.Fatalf("missing phase-1 dep-update:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: dep-update example.com/leaf/cmd <- example.com/leaf") {
		t.Fatalf("missing phase-2 dep-update:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: go mod tidy  (local git)") {
		t.Fatalf("missing local-git tidy:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: git commit -m 'dep: example.com/leaf v0.0.1 -> v0.0.2'") {
		t.Fatalf("missing phase-2 commit subject:\n%s", stdout)
	}
	if !strings.Contains(stdout, "skipped ("+phase2EmptyNote+")") {
		t.Fatalf("empty phase-2 repo missing skip note:\n%s", stdout)
	}
	if strings.Contains(stdout, "example.com/sdk") {
		t.Fatalf("catch-up pin leaked into dry-run:\n%s", stdout)
	}
	if !strings.Contains(stdout, "would: reinstall-local\n") {
		t.Fatalf("missing phase-1 reinstall-local:\n%s", stdout)
	}
	if strings.Contains(stdout, "would: reinstall-local ") {
		t.Fatalf("reinstall-local should not repeat the path:\n%s", stdout)
	}
	idxSkip := strings.Index(stdout, "skipped ("+phase2EmptyNote+")")
	if idxSkip < 0 {
		t.Fatalf("missing phase-2 skip:\n%s", stdout)
	}
	p12, shipOut := stdout[:idxSkip], stdout[idxSkip:]
	if strings.Contains(p12, "would: push") || strings.Contains(p12, "would: sync") {
		t.Fatalf("push/sync leaked into phase-1/2:\n%s", p12)
	}
	if strings.Contains(shipOut, "reinstall") {
		t.Fatalf("reinstall leaked into ship:\n%s", shipOut)
	}
	if !strings.Contains(shipOut, "would: push\n") || !strings.Contains(shipOut, "would: sync\n") {
		t.Fatalf("missing grouped ship push/sync:\n%s", shipOut)
	}
	if strings.Contains(shipOut, "would: push ") || strings.Contains(shipOut, "would: sync ") {
		t.Fatalf("ship kinds should not repeat the path:\n%s", shipOut)
	}
	// Phase 2 empty root sits after leaf (peel order), not as a module-path header.
	idxLeaf := strings.Index(stdout, "\n      leaf\n")
	if idxLeaf < 0 || idxSkip < idxLeaf {
		t.Fatalf("phase-2 repo order unexpected:\n%s", stdout)
	}
}

func TestSplitJobPlanActionsOmitsCatchUpAndSplitsShip(t *testing.T) {
	t.Parallel()
	snap := sampleJobSnap()
	flags := UnwindFlags{MergeBack: true, TagNext: true, GenCommitMsg: true, Push: true, Sync: true, ReinstallLocal: true}
	job := BuildJobPlan(snap, flags)
	p1, p2, ship := splitJobPlanActions(job)
	ids := func(list []*Action) []string {
		var out []string
		for _, a := range list {
			out = append(out, a.ID)
		}
		return out
	}
	for _, id := range ids(p1) {
		if strings.Contains(id, "example.com/sdk") {
			t.Fatalf("phase1 still has catch-up pin %s: %v", id, ids(p1))
		}
		if strings.HasPrefix(id, "push:") || strings.HasPrefix(id, "sync:") {
			t.Fatalf("ship action in phase1: %s", id)
		}
	}
	if findIn(p1, "pin:example.com/root<example.com/leaf") == nil {
		t.Fatalf("phase1 missing propagate pin, %v", ids(p1))
	}
	if findIn(p1, "reinstall:leaf") == nil {
		t.Fatalf("phase1 missing reinstall, %v", ids(p1))
	}
	if findIn(p2, "pin:example.com/leaf/cmd<example.com/leaf") == nil {
		t.Fatalf("phase2 missing intra pin, %v", ids(p2))
	}
	if findIn(ship, "push:leaf") == nil && findIn(ship, "push:root") == nil {
		t.Fatalf("ship missing push, %v", ids(ship))
	}
	if findIn(ship, "reinstall:leaf") != nil {
		t.Fatalf("ship must not include reinstall, %v", ids(ship))
	}
}

func TestCheckoutOfFollowsLand(t *testing.T) {
	t.Parallel()
	r := &jobRunner{
		byLabel: map[string]StackMember{
			"app": {Path: "/wt", MainRepo: "/main", Label: "app", Linked: true},
		},
		checkoutByLane: map[string]string{"app": "/wt"},
		peeled:         map[string]bool{},
	}
	if got := r.checkoutOf("app"); got != "/wt" {
		t.Fatalf("before land checkout=%s want /wt", got)
	}
	r.checkoutByLane["app"] = "/main"
	if got := r.checkoutOf("app"); got != "/main" {
		t.Fatalf("after land checkout=%s want /main", got)
	}
}

func TestApplyMergeBackUnlinkedUsesMain(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	st := newStageWriter(4, false, false)
	st.outW, st.errW = &out, &out
	r := &jobRunner{
		byLabel: map[string]StackMember{
			"app": {Path: "/main", MainRepo: "/main", Label: "app", Linked: false},
		},
		checkoutByLane: map[string]string{"app": "/wt-stale"},
		st:             st,
		stats:          &UnwindApplyStats{},
	}
	if err := r.applyMergeBack("app", false); err != nil {
		t.Fatal(err)
	}
	if got := r.checkoutOf("app"); got != "/main" {
		t.Fatalf("unlinked merge-back checkout=%s want /main", got)
	}
}

func findIn(list []*Action, id string) *Action {
	for _, a := range list {
		if a != nil && a.ID == id {
			return a
		}
	}
	return nil
}
