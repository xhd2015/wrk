package workops

import "github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"

// StatusReport is a structured checkout report for Status.
type StatusReport struct {
	MainPath     string
	CheckoutPath string
	Branch       string
	HeadShort    string
	IsWorktree   bool
	// Optional dirty counts when cheap to compute.
	Added   int
	Changed int
	Renamed int
	Deleted int
}

// Project is one registered main repository entry from wrk home.
type Project struct {
	Path      string
	OriginURL string // optional; may be empty when not stored
}

// MergeBackOptions configures landing a linked worktree into main.
// Remove is always false for this API (worktree is kept).
type MergeBackOptions struct {
	WorktreeDir string
	Sync        bool
	DryRun      bool
	WrkHome     string
	// Confirm is called when the land plan needs user confirmation.
	// nil means auto-confirm (library / hermetic tests).
	// CLI injects worktree.PromptConfirmPlan.
	Confirm func(plan worktree.MergeBackPlan) (bool, error)
}

// MergeBackResult describes the outcome of a land (for CLI composition).
type MergeBackResult struct {
	SourcePath string
	TargetPath string
	Branch     string
	Relation   string
	// Action: "noop", "merged", "rebased-and-merged", "dry-run", "aborted", …
	Action  string
	Message string
}

// TagNextOptions configures planning/applying the next release tag(s).
type TagNextOptions struct {
	Checkout string
	DryRun   bool
	// HeadRef is the commit or symbolic ref to plan/apply against.
	// Empty defaults to "HEAD" (main tip after resolve).
	HeadRef string
}

// PushOptions configures pushing the current branch and optional tags.
type PushOptions struct {
	Checkout string
	DryRun   bool
	// Force uses git push --force-with-lease for the branch only.
	// Tags are always pushed without force.
	Force bool
	Tags  []string
}
