package unwind

import (
	"strings"
)

// JobStep is one preview/run action.
type JobStep struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
	Detail string `json:"detail,omitempty"`
	Why    string `json:"why,omitempty"`
}

// JobEpoch is a named group of steps matching ApplyUnwind epochs.
type JobEpoch struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Steps []JobStep `json:"steps"`
}

// JobFlags is the subset of UnwindFlags that shapes a job preview.
type JobFlags struct {
	TagNext        bool   `json:"tag_next"`
	Push           bool   `json:"push"`
	Force          bool   `json:"force"`
	Done           bool   `json:"done"`
	MergeBack      bool   `json:"merge_back"`
	Sync           bool   `json:"sync"`
	ReinstallLocal bool   `json:"reinstall_local"`
	AddAll         bool   `json:"add_all"`
	GenCommitMsg   bool   `json:"gen_commit_msg"`
	Commit         bool   `json:"commit"` // --commit in GenCommitArgs (required with gen-commit-msg)
	NoVerify       bool   `json:"no_verify"`
	AgentRunner    string `json:"agent_runner,omitempty"` // display/passthrough from CLI peel
	Cleanup        bool   `json:"cleanup"`                // include latest-drift / drop-replace pins
}

// JobPlan is the action preview for --unwind --web (and dry-run epochs).
type JobPlan struct {
	WorkDir     string             `json:"work_dir"`
	Flags       JobFlags           `json:"flags"`
	Graph       *UnwindGraphReport `json:"graph,omitempty"`
	ActionGraph *ActionGraph       `json:"action_graph,omitempty"`
	// Phases is snapshot, repos (cross-repo), modules (intra), ship (push+sync).
	// Apply and --dry-run walk this DAG. Web is the execution model.
	Phases    []JobPhase `json:"phases,omitempty"`
	Epochs    []JobEpoch `json:"epochs"`
	Blockers  []string   `json:"blockers"`
	CanRun    bool       `json:"can_run"`
	ApplyHint string     `json:"apply_hint,omitempty"`
}

func flagsFromUnwind(f UnwindFlags) JobFlags {
	return JobFlags{
		TagNext:        f.TagNext,
		Push:           f.Push,
		Force:          f.Force,
		Done:           f.Done,
		MergeBack:      f.MergeBack,
		Sync:           f.Sync,
		ReinstallLocal: f.ReinstallLocal,
		AddAll:         f.AddAll || genArgsHasFlag(f.GenCommitArgs, "--add-all"),
		GenCommitMsg:   f.GenCommitMsg,
		Commit:         genArgsHasFlag(f.GenCommitArgs, "--commit"),
		NoVerify:       genArgsHasFlag(f.GenCommitArgs, "--no-verify"),
		AgentRunner:    genArgsFlagValue(f.GenCommitArgs, "--agent-runner"),
		Cleanup:        f.Cleanup,
	}
}

// FlagsFromJob maps checkbox JSON onto UnwindFlags.
// baseGenArgs supplies non-toggle peeled values (e.g. --agent-runner=…).
func FlagsFromJob(j JobFlags, baseGenArgs []string) UnwindFlags {
	return unwindFlagsFromJob(j, baseGenArgs)
}

func unwindFlagsFromJob(j JobFlags, baseGenArgs []string) UnwindFlags {
	return UnwindFlags{
		TagNext:        j.TagNext,
		Push:           j.Push,
		Force:          j.Force,
		Done:           j.Done,
		MergeBack:      j.MergeBack,
		Sync:           j.Sync,
		ReinstallLocal: j.ReinstallLocal,
		AddAll:         j.AddAll,
		GenCommitMsg:   j.GenCommitMsg,
		GenCommitArgs:  jobGenCommitArgs(j, baseGenArgs),
		Cleanup:        j.Cleanup,
	}
}

// jobGenCommitArgs keeps non-bool peeled gen-commit flags from base and
// overlays --commit / --add-all / --no-verify from JobFlags checkboxes.
func jobGenCommitArgs(j JobFlags, base []string) []string {
	out := stripGenCommitBoolFlags(base)
	if j.AgentRunner != "" && !genArgsHasFlag(out, "--agent-runner") {
		out = append(out, "--agent-runner="+j.AgentRunner)
	}
	if j.Commit {
		out = append(out, "--commit")
	}
	if j.AddAll {
		out = append(out, "--add-all")
	}
	if j.NoVerify {
		out = append(out, "--no-verify")
	}
	return out
}

func stripGenCommitBoolFlags(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		name := a
		if i := strings.IndexByte(a, '='); i >= 0 {
			name = a[:i]
		}
		switch name {
		case "--commit", "--add-all", "--no-verify":
			continue
		default:
			out = append(out, a)
		}
	}
	return out
}

func genArgsFlagValue(genArgs []string, flag string) string {
	for i, a := range genArgs {
		name := a
		val := ""
		if j := strings.IndexByte(a, '='); j >= 0 {
			name = a[:j]
			val = a[j+1:]
		}
		if name != flag {
			continue
		}
		if val != "" {
			return val
		}
		if i+1 < len(genArgs) && !strings.HasPrefix(genArgs[i+1], "-") {
			return genArgs[i+1]
		}
		return ""
	}
	return ""
}

// BuildJobPlan builds a preview from a cascade-capable snapshot and flags.
func BuildJobPlan(snap *Snapshot, flags UnwindFlags) *JobPlan {
	out := &JobPlan{
		Flags: flagsFromUnwind(flags),
	}
	if snap == nil {
		out.Blockers = []string{"unwind: no snapshot"}
		return out
	}
	out.WorkDir = snap.WorkDir
	plan := snap.Peel
	if plan == nil {
		plan = &UnwindPlan{}
	}

	// Re-attach NextTag from worktree tip (staged / add-all) using current flags.
	snap = applyTipAwareTags(snap, flags)

	opts := ActionGraphOpts{Cleanup: flags.Cleanup}
	phases := buildJobPhases(snap, flags, opts)
	out.Phases = phases
	// action_graph stays phase-1 (repos) for older clients / epochs bridge.
	if ph := phaseByID(phases, "repos"); ph != nil && ph.ActionGraph != nil {
		out.ActionGraph = ph.ActionGraph
	} else {
		out.ActionGraph = BuildActionGraph(snap, flags, opts)
	}
	out.Epochs = epochsFromActionGraph(out.ActionGraph)

	if err := ValidateUnwindFlags(plan, flags); err != nil {
		out.Blockers = []string{err.Error()}
	}
	out.CanRun = len(out.Blockers) == 0
	out.ApplyHint = buildApplyHintRemaining(plan, flags)
	if out.ActionGraph != nil && out.ActionGraph.FilterNote != "" && out.ApplyHint == "" {
		out.ApplyHint = out.ActionGraph.FilterNote
	} else if out.ActionGraph != nil && out.ActionGraph.FilterNote != "" {
		out.ApplyHint = out.ApplyHint + "; " + out.ActionGraph.FilterNote
	}
	if g, err := graphFromSnapshot(snap); err == nil {
		out.Graph = g
	}
	return out
}
