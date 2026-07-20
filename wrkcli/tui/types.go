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

// StageRunState is the per-stage run indicator for RUN ALL / single-stage ops.
type StageRunState string

const (
	StageIdle    StageRunState = "idle"
	StageQueued  StageRunState = "queued"
	StageRunning StageRunState = "running"
	StageOK      StageRunState = "ok"
	StageError   StageRunState = "error"
	StageSkipped StageRunState = "skipped"
)

// StageEvent is one stage transition during a phase-aware pipeline run.
// StageID values: add-changes | gen-commit-msg | commit | merge-back | done |
// sync | tag-next | push | reinstall-local.
// Kind values: start | ok | error | skipped.
type StageEvent struct {
	StageID string
	Kind    string // start | ok | error | skipped
	Err     string // if error
	Result  string // brief one-line for any stage (e.g. "staged", "planned")
	Subject string // gen-commit subject (first line of message)
	Body    string // remaining body (optional; may go to Log)
}

// LogFunc is a TUI-safe logger used while the dashboard owns the terminal.
// Lines must be delivered into the model Log ring (via tea msgs), never written
// with fmt.Print / log.Printf to the real tty (that corrupts Bubble Tea frames).
// stage may be empty for pipeline-level notes.
type LogFunc func(stage, line string)

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
	// RunCompose executes compose for the given recipe.
	// Legacy bridge: TUI may temporarily redirect process stdout/stderr to feed the Log panel.
	// Prefer RunPipeline + LogFunc for structured stage output. Used for single-stage runs
	// and as fallback when RunPipeline is nil.
	RunCompose func(workDir string, r Recipe) error
	// RunPipeline runs the recipe as ordered stages.
	// emit delivers stage state transitions; log delivers user-visible lines for the Log panel
	// (TUI-safe — do not fmt/log to the real tty from pipeline code).
	// If nil, fall back to RunCompose once.
	RunPipeline func(workDir string, r Recipe, emit func(StageEvent), log LogFunc) error
	// ComposeArgv builds CLI tokens for preview / status from a recipe.
	ComposeArgv func(r Recipe) []string

	// StagePreview returns a brief one-line preview for stageID (e.g. "add-changes")
	// plus optional log lines (captured tool stderr/diagnostics) for the Log panel.
	// Preview text may be empty on soft failure; Logs must still be delivered — never
	// discarded and never printed to the real tty. Called from background tea.Cmds
	// after first paint — must not block Init.
	StagePreview func(workDir string, stageID string) StagePreviewResult
}

// StagePreviewResult is one async stage preview probe outcome.
type StagePreviewResult struct {
	// Preview is the one-line stage meta (may be empty).
	Preview string
	// Logs are normal Log panel lines (e.g. captured "fatal: no upstream…").
	// Never discarded; never written to the real terminal by the probe path.
	Logs []string
}
