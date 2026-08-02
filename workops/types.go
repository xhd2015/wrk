package workops

// StatusReport is a structured checkout report for Status.
type StatusReport struct {
	MainPath     string
	CheckoutPath string
	Branch       string
	HeadShort    string
	IsWorktree   bool
	// Optional dirty counts when cheap to compute.
	Added    int
	Changed  int
	Renamed  int
	Deleted  int
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
}

// TagNextOptions configures planning/applying the next release tag(s).
type TagNextOptions struct {
	Checkout string
	DryRun   bool
}

// PushOptions configures pushing the current branch and optional tags.
type PushOptions struct {
	Checkout string
	DryRun   bool
	Tags     []string
}
