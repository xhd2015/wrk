package unwind

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSplitPeelOrderB1UsesProvidedCascade(t *testing.T) {
	t.Parallel()
	members := []StackMember{
		{Path: "/tmp/leaf", Label: "leaf", Dirty: true},
		{Path: "/tmp/root", Label: "root", Dirty: true},
	}
	peel := []string{"leaf", "root"}
	nodes := []UnwindGraphModuleNode{
		{Path: "example.com/leaf", RepoLabel: "leaf"},
		{Path: "example.com/root", RepoLabel: "root"},
	}
	edges := []UnwindGraphModuleEdge{
		{From: "example.com/root", To: "example.com/leaf", Kind: "require"},
	}
	cascade := &UnwindCascadePlan{Steps: []UnwindCascadeStep{
		{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
		{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
	}}
	early, deferred := splitPeelOrderB1(peel, members, cascade, nodes, edges)
	if len(early) != 1 || early[0] != "leaf" {
		t.Fatalf("early=%v want [leaf]", early)
	}
	if len(deferred) != 1 || deferred[0] != "root" {
		t.Fatalf("deferred=%v want [root]", deferred)
	}
}

func TestSplitPeelOrderB1EmptyCascadeAllEarly(t *testing.T) {
	t.Parallel()
	peel := []string{"a", "b"}
	early, deferred := splitPeelOrderB1(peel, []StackMember{{Label: "a"}, {Label: "b"}}, nil, nil, nil)
	if len(deferred) != 0 {
		t.Fatalf("deferred=%v", deferred)
	}
	if len(early) != 2 || early[0] != "a" || early[1] != "b" {
		t.Fatalf("early=%v", early)
	}
}

func TestFormatUnwindDryRunEpochHeaders(t *testing.T) {
	t.Parallel()
	plan := &UnwindPlan{PeelOrder: []string{"leaf", "root"}}
	members := []StackMember{
		{Path: "/tmp/leaf", MainRepo: "/tmp/leaf", Label: "leaf", Dirty: true},
		{Path: "/tmp/root", MainRepo: "/tmp/root", Label: "root", Dirty: true},
	}
	cascade := &UnwindCascadePlan{Steps: []UnwindCascadeStep{
		{Kind: CascadeTagNext, ModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
		{Kind: CascadePin, ModulePath: "example.com/root", DepModulePath: "example.com/leaf", TagOrVersion: "v0.0.2"},
	}}
	nodes := []UnwindGraphModuleNode{
		{Path: "example.com/leaf", RepoLabel: "leaf"},
		{Path: "example.com/root", RepoLabel: "root"},
	}
	edges := []UnwindGraphModuleEdge{
		{From: "example.com/root", To: "example.com/leaf", Kind: "require"},
	}
	out := formatUnwindDryRun(plan, members, "/tmp", UnwindFlags{TagNext: true, Push: true}, cascade, nodes, edges)
	for _, want := range []string{
		"---- early peels ----",
		"would: peel",
		"---- cascade ----",
		"would: tag-next example.com/leaf @ v0.0.2",
		"would: pin example.com/root <- example.com/leaf @ v0.0.2",
		"---- deferred peels ----",
		"---- ship ----",
		"would: push branch and created tag",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in\n%s", want, out)
		}
	}
}

func TestBuildJobPlanBlockers(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/a", Label: "a", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"a"}, NeedsLand: true},
	}
	job := BuildJobPlan(snap, UnwindFlags{})
	if job.CanRun {
		t.Fatal("expected CanRun false without land flags")
	}
	if len(job.Blockers) == 0 {
		t.Fatal("expected blockers")
	}
	job2 := BuildJobPlan(snap, UnwindFlags{MergeBack: true})
	if !job2.CanRun {
		t.Fatalf("expected CanRun after merge-back, blockers=%v", job2.Blockers)
	}
}

func TestBuildJobPlanMergeBackAndDoneKinds(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/a", Label: "a", Dirty: true, Linked: true},
			{Path: "/tmp/b", Label: "b", Dirty: true, Linked: false},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"a", "b"}, NeedsLand: true},
	}
	mb := BuildJobPlan(snap, UnwindFlags{MergeBack: true})
	kinds := jobStepKinds(mb)
	if !containsKind(kinds, "merge-back") {
		t.Fatalf("merge-back plan kinds=%v want merge-back", kinds)
	}
	if containsKind(kinds, "land") || containsKind(kinds, "done") {
		t.Fatalf("merge-back plan must not emit land/done: %v", kinds)
	}
	done := BuildJobPlan(snap, UnwindFlags{Done: true})
	kindsDone := jobStepKinds(done)
	if !containsKind(kindsDone, "done") {
		t.Fatalf("done plan kinds=%v want done", kindsDone)
	}
	if containsKind(kindsDone, "land") || containsKind(kindsDone, "merge-back") {
		t.Fatalf("done plan must not emit land/merge-back: %v", kindsDone)
	}
	// Main-only peel has no merge-back/done step.
	for _, ep := range mb.Epochs {
		for _, st := range ep.Steps {
			if st.Target == "b" && (st.Kind == "merge-back" || st.Kind == "done") {
				t.Fatalf("main checkout must not get %s step", st.Kind)
			}
		}
	}
}

func jobStepKinds(job *JobPlan) []string {
	var out []string
	for _, ep := range job.Epochs {
		for _, st := range ep.Steps {
			out = append(out, st.Kind)
		}
	}
	return out
}

func containsKind(kinds []string, want string) bool {
	for _, k := range kinds {
		if k == want {
			return true
		}
	}
	return false
}

func TestFlagsFromJobPreservesAgentRunnerAndCommitToggles(t *testing.T) {
	t.Parallel()
	base := []string{"--agent-runner=commandcode", "--commit", "--add-all"}
	fromCLI := UnwindFlags{
		GenCommitMsg:  true,
		MergeBack:     true,
		GenCommitArgs: base,
	}
	jf := flagsFromUnwind(fromCLI)
	if !jf.GenCommitMsg || !jf.Commit || !jf.AddAll {
		t.Fatalf("flagsFromUnwind=%+v", jf)
	}
	if jf.AgentRunner != "commandcode" {
		t.Fatalf("agent_runner=%q", jf.AgentRunner)
	}
	// Uncheck commit in UI; keep agent-runner.
	jf.Commit = false
	jf.AddAll = true
	out := FlagsFromJob(jf, base)
	if !out.GenCommitMsg || !out.AddAll {
		t.Fatalf("out=%+v", out)
	}
	if genArgsHasFlag(out.GenCommitArgs, "--commit") {
		t.Fatalf("commit should be cleared: %v", out.GenCommitArgs)
	}
	if !genArgsHasFlag(out.GenCommitArgs, "--add-all") {
		t.Fatalf("add-all should remain: %v", out.GenCommitArgs)
	}
	if genArgsFlagValue(out.GenCommitArgs, "--agent-runner") != "commandcode" {
		t.Fatalf("agent-runner lost: %v", out.GenCommitArgs)
	}
}

func TestBuildJobPlanApplyHintRemaining(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/a", Label: "a", Dirty: true, Linked: true},
		}},
		Peel: &UnwindPlan{
			PeelOrder:       []string{"a"},
			NeedsLand:       true,
			HasPendingEdges: true,
		},
	}
	full := BuildJobPlan(snap, UnwindFlags{})
	if full.ApplyHint != "apply would need --merge-back --tag-next --push" {
		t.Fatalf("full hint=%q", full.ApplyHint)
	}
	partial := BuildJobPlan(snap, UnwindFlags{MergeBack: true, TagNext: true})
	if partial.ApplyHint != "apply would need --push" {
		t.Fatalf("partial hint=%q", partial.ApplyHint)
	}
	done := BuildJobPlan(snap, UnwindFlags{Done: true, TagNext: true, Push: true})
	if done.ApplyHint != "" {
		t.Fatalf("satisfied hint=%q want empty", done.ApplyHint)
	}
}

func TestBuildJobPlanGraphJSONTags(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		WorkDir: "/tmp/stack",
		Inv: StackInventory{Members: []StackMember{
			{Path: "/tmp/stack", Label: "stack", Dirty: true},
		}},
		Peel: &UnwindPlan{PeelOrder: []string{"stack"}},
	}
	job := BuildJobPlan(snap, UnwindFlags{})
	raw, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	g, ok := decoded["graph"].(map[string]any)
	if !ok {
		t.Fatalf("graph missing or wrong type: %T", decoded["graph"])
	}
	if _, ok := g["summary"].(map[string]any); !ok {
		t.Fatalf("graph.summary missing: %#v", g)
	}
	repos, ok := g["repos"].(map[string]any)
	if !ok {
		t.Fatalf("graph.repos missing: %#v", g)
	}
	if _, ok := repos["peel_order"]; !ok {
		t.Fatalf("graph.repos.peel_order missing: %#v", repos)
	}
}
