// Package tui implements the interactive Bubble Tea dashboard for bare `wrk`.
// Dependencies (git, compose, dirt gates) are injected via RunDashboardOpts so
// this package never imports parent github.com/xhd2015/wrk/wrkcli.
package tui

// Recipe holds stage selection for RUN / single-phase compose.
type Recipe struct {
	AddAll         bool
	GenCommitMsg   bool
	Commit         bool
	AgentRunner    string
	Done           bool
	MergeBack      bool
	Sync           bool
	TagNext        bool
	Push           bool
	ReinstallLocal bool
	DryRun         bool
}

// RunDashboardOpts configures the interactive dashboard TUI and injects deps.
type RunDashboardOpts struct {
	WorkDir string
	Status  string

	// HasAddableDirt reports whether workDir has unstaged/untracked dirt.
	HasAddableDirt func(workDir string) bool
	// IsMainCheckout reports whether workDir is the primary (non-linked) checkout.
	IsMainCheckout func(workDir string) bool
	// GitAddAll stages all changes (git add -A) in workDir.
	GitAddAll func(workDir string) error
	// RunCompose executes compose for the given recipe (stdio may be redirected by caller of RunDashboard ops).
	RunCompose func(workDir string, r Recipe) error
	// ComposeArgv builds CLI tokens for preview / status from a recipe.
	ComposeArgv func(r Recipe) []string
}
