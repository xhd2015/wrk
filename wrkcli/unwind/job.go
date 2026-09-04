package unwind

import (
	"fmt"
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
}

// JobPlan is the action preview for --unwind --web (and dry-run epochs).
type JobPlan struct {
	WorkDir   string             `json:"work_dir"`
	Flags     JobFlags           `json:"flags"`
	Graph     *UnwindGraphReport `json:"graph,omitempty"`
	Epochs    []JobEpoch         `json:"epochs"`
	Blockers  []string           `json:"blockers"`
	CanRun    bool               `json:"can_run"`
	ApplyHint string             `json:"apply_hint,omitempty"`
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
	members := snap.Inv.Members
	byLabel := pickPeelMembersByLabel(members)
	cascade := snap.Cascade
	if cascade == nil {
		cascade = &UnwindCascadePlan{}
	}

	var early, deferred []string
	if flags.TagNext {
		early, deferred = splitPeelOrderB1(plan.PeelOrder, members, cascade, snap.ModuleNodes, snap.ModuleEdges)
	} else {
		early = append([]string(nil), plan.PeelOrder...)
	}

	peelSteps := func(labels []string) []JobStep {
		var steps []JobStep
		for _, label := range labels {
			display := label
			m, ok := byLabel[label]
			if ok {
				display = peelDisplayPath(snap.WorkDir, m.Path)
			}
			why := "dirty stack checkout"
			steps = append(steps, JobStep{Kind: "peel", Target: display, Why: why})
			if flags.GenCommitMsg {
				steps = append(steps, JobStep{Kind: "gen-commit", Target: display, Why: "unstaged or staged feature work"})
			}
			if ok && m.Linked && (flags.Done || flags.MergeBack) {
				if flags.Done {
					steps = append(steps, JobStep{
						Kind: "done", Target: display,
						Detail: "merge-back and remove worktree", Why: "linked worktree",
					})
				} else {
					steps = append(steps, JobStep{
						Kind: "merge-back", Target: display,
						Detail: "keep worktree", Why: "linked worktree",
					})
				}
			}
		}
		return steps
	}

	if flags.TagNext {
		if steps := peelSteps(early); len(steps) > 0 {
			out.Epochs = append(out.Epochs, JobEpoch{ID: "early-peel", Title: "early peels", Steps: steps})
		}
		var cas []JobStep
		for _, s := range cascade.Steps {
			switch s.Kind {
			case CascadeTagNext:
				cas = append(cas, JobStep{
					Kind:   "tag-next",
					Target: s.ModulePath,
					Detail: s.TagOrVersion,
					Why:    "owned changes",
				})
			case CascadePin:
				cas = append(cas, JobStep{
					Kind:   "pin",
					Target: s.ModulePath,
					Detail: fmt.Sprintf("<- %s @ %s", s.DepModulePath, s.TagOrVersion),
					Why:    "require drift or droppable replace",
				})
			}
		}
		if len(cas) > 0 {
			out.Epochs = append(out.Epochs, JobEpoch{ID: "cascade", Title: "cascade", Steps: cas})
		}
		if steps := peelSteps(deferred); len(steps) > 0 {
			out.Epochs = append(out.Epochs, JobEpoch{ID: "deferred-peel", Title: "deferred peels", Steps: steps})
		}
	} else if steps := peelSteps(plan.PeelOrder); len(steps) > 0 {
		out.Epochs = append(out.Epochs, JobEpoch{ID: "peel", Title: "peels", Steps: steps})
	}

	var ship []JobStep
	if flags.Push {
		ship = append(ship, JobStep{Kind: "push", Target: "touched mains", Why: "publish branch and tags"})
	}
	if flags.Sync {
		ship = append(ship, JobStep{Kind: "sync", Target: "linked worktrees", Why: "fast-forward after merge-back"})
	}
	if len(ship) > 0 {
		out.Epochs = append(out.Epochs, JobEpoch{ID: "ship", Title: "ship", Steps: ship})
	}
	if flags.ReinstallLocal {
		out.Epochs = append(out.Epochs, JobEpoch{
			ID:    "reinstall",
			Title: "reinstall",
			Steps: []JobStep{{Kind: "reinstall-local", Target: "touched mains", Why: "local binaries"}},
		})
	}

	if err := ValidateUnwindFlags(plan, flags); err != nil {
		out.Blockers = []string{err.Error()}
	}
	out.CanRun = len(out.Blockers) == 0
	out.ApplyHint = buildApplyHintRemaining(plan, flags)
	if g, err := graphFromSnapshot(snap); err == nil {
		out.Graph = g
	}
	return out
}
