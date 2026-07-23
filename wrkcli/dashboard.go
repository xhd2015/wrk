package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// Dashboard glyphs (fine-grained; never [x]/[X]).
const (
	dashGlyphOn       = "[•]"
	dashGlyphOff      = "[ ]"
	dashGlyphDisabled = "[-]"
)

// Hermetic interactive hooks (doctest / non-TTY forced actions).
const (
	envDashboardAction         = "WRK_DASHBOARD_ACTION"
	envDashboardDryRun         = "WRK_DASHBOARD_DRY_RUN"
	envDashboardComposeArgvLog = "WRK_DASHBOARD_COMPOSE_ARGV_LOG"
	envDashboardToggles        = "WRK_DASHBOARD_TOGGLES"
)

// dashboardStage is one selectable row in the static dashboard View.
type dashboardStage struct {
	// Glyph is one of [•] / [ ] / [-].
	Glyph string
	// Label is the human-readable stage name (e.g. "add changes", "DONE").
	Label string
	// Nested optional secondary line(s), e.g. agent-runner under gen-commit-msg.
	Nested []string
}

// dashboardSection groups stages under Pre / Main / After.
type dashboardSection struct {
	Title  string
	Stages []dashboardStage
}

// dashboardModel is the pure snapshot model for bare `wrk` (static View).
type dashboardModel struct {
	// Sections in display order: Pre, Main, After.
	Sections []dashboardSection
}

// buildDashboardModel inspects workDir and builds the default stage snapshot.
// Dirty unstaged/untracked → add changes enabled ([•]); clean → disabled ([-]).
// Main: MERGE BACK first and default (safer than DONE).
func buildDashboardModel(workDir string) dashboardModel {
	addGlyph := dashGlyphDisabled
	if dashboardHasAddableDirt(workDir) {
		addGlyph = dashGlyphOn
	}

	// MERGE BACK default selected; both disabled on main checkout.
	doneGlyph := dashGlyphOff
	mergeGlyph := dashGlyphOn
	if dashboardIsMainCheckout(workDir) {
		doneGlyph = dashGlyphDisabled
		mergeGlyph = dashGlyphDisabled
	}

	return dashboardModel{
		Sections: []dashboardSection{
			{
				Title: "Pre",
				Stages: []dashboardStage{
					{Glyph: addGlyph, Label: "add changes"},
					{
						Glyph: dashGlyphOn,
						Label: "gen-commit-msg",
						Nested: []string{
							"  agent-runner: commandcode",
						},
					},
					{Glyph: dashGlyphOn, Label: "commit"},
				},
			},
			{
				Title: "Main",
				Stages: []dashboardStage{
					{Glyph: mergeGlyph, Label: "MERGE BACK"},
					{Glyph: doneGlyph, Label: "DONE"},
				},
			},
			{
				Title: "After",
				Stages: []dashboardStage{
					{Glyph: dashGlyphOn, Label: "sync"},
					{Glyph: dashGlyphOn, Label: "tag-next"},
					{Glyph: dashGlyphOn, Label: "push"},
					{Glyph: dashGlyphOn, Label: "reinstall-local"},
				},
			},
		},
	}
}

// renderDashboardView formats the compact non-TTY snapshot (no ANSI, no create hint).
func renderDashboardView(m dashboardModel) string {
	const w = 60
	var b strings.Builder
	line := func(inner string) {
		innerW := w - 2
		if len(inner) > innerW {
			inner = inner[:innerW]
		}
		b.WriteString("│")
		b.WriteString(inner)
		if pad := innerW - len(inner); pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString("│\n")
	}
	section := func(name string) {
		body := "─ " + name + " "
		rest := w - 2 - len(body)
		if rest < 0 {
			rest = 0
		}
		b.WriteString("├")
		b.WriteString(body)
		b.WriteString(strings.Repeat("─", rest))
		b.WriteString("┤\n")
	}

	title := "─ wrk dashboard (non-interactive) "
	rest := w - 2 - len(title)
	if rest < 0 {
		rest = 0
	}
	b.WriteString("┌")
	b.WriteString(title)
	b.WriteString(strings.Repeat("─", rest))
	b.WriteString("┐\n")
	line(" status  ready")

	for _, sec := range m.Sections {
		section(sec.Title)
		for _, st := range sec.Stages {
			row := " " + st.Glyph + " " + st.Label
			if len(st.Nested) > 0 {
				// Inline nested (e.g. agent-runner) for compactness
				row += "  " + strings.TrimSpace(st.Nested[0])
			}
			if sec.Title != "Main" {
				// Right-align placeholder Run rail for visual parity
				run := "[ Run ]"
				for len(row)+len(run) < w-2 {
					row += " "
				}
				if len(row)+len(run) > w-2 {
					row = row[:w-2-len(run)]
				}
				row += run
			}
			line(row)
		}
	}

	preview := dashboardRecipe{
		addAll:         m.Sections[0].Stages[0].Glyph == dashGlyphOn,
		genCommitMsg:   true,
		commit:         true,
		agentRunner:    "commandcode",
		mergeBack:      m.Sections[1].Stages[0].Glyph == dashGlyphOn,
		done:           m.Sections[1].Stages[1].Glyph == dashGlyphOn,
		sync:           true,
		tagNext:        true,
		push:           true,
		reinstallLocal: true,
	}
	if m.Sections[1].Stages[0].Glyph == dashGlyphDisabled {
		preview.mergeBack = false
		preview.done = false
	}
	section("Batch")
	argv := " would run: " + strings.Join(composeArgvFromRecipe(preview), " ")
	for len(argv) > 0 {
		chunk := argv
		if len(chunk) > w-2 {
			chunk = argv[:w-2]
			argv = argv[w-2:]
		} else {
			argv = ""
		}
		line(chunk)
	}
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", w-2))
	b.WriteString("┘\n")
	return b.String()
}

// runDashboard is the bare `wrk` entry: static snapshot, hermetic ACTION, or TTY TUI.
func runDashboard(workDir string, ctx *invocationContext) error {
	action := strings.TrimSpace(os.Getenv(envDashboardAction))
	stdinTTY := term.IsTerminal(int(os.Stdin.Fd()))

	// Hermetic ACTION forces the interactive path even without a TTY.
	if action != "" {
		return applyDashboardAction(workDir, ctx, action)
	}

	// Real TTY: web-like Bubble Tea; stay open across runs until CANCEL.
	if stdinTTY {
		return runDashboardTeaLoop(workDir, ctx)
	}

	// Non-TTY, no ACTION → static snapshot only.
	m := buildDashboardModel(workDir)
	fmt.Fprint(cliStdout(), renderDashboardView(m))
	return nil
}

// applyDashboardAction handles WRK_DASHBOARD_ACTION=cancel|run-done|run-merge-back.
func applyDashboardAction(workDir string, ctx *invocationContext, action string) error {
	switch action {
	case "cancel":
		// Exit 0; no compose; leave argv log empty; keep event command dashboard.
		fmt.Fprint(cliStdout(), renderDashboardView(buildDashboardModel(workDir)))
		return nil
	case "run-done":
		return runDashboardCompose(workDir, ctx, true /* done */)
	case "run-merge-back":
		return runDashboardCompose(workDir, ctx, false /* merge-back */)
	default:
		return fmt.Errorf("wrk: unknown %s=%q (want cancel|run-done|run-merge-back)", envDashboardAction, action)
	}
}

// dashboardRecipe holds stage selection for RUN compose.
type dashboardRecipe struct {
	addAll         bool
	genCommitMsg   bool
	commit         bool
	agentRunner    string
	done           bool
	mergeBack      bool
	sync           bool
	tagNext        bool
	push           bool
	reinstallLocal bool
	dryRun         bool
}

// defaultDashboardRecipe builds DONE or MERGE BACK defaults for RUN.
// primaryDone false → MERGE BACK (safer default).
func defaultDashboardRecipe(workDir string, primaryDone bool) dashboardRecipe {
	r := dashboardRecipe{
		addAll:         dashboardHasAddableDirt(workDir),
		genCommitMsg:   true,
		commit:         true,
		agentRunner:    "commandcode",
		done:           primaryDone,
		mergeBack:      !primaryDone,
		sync:           true,
		tagNext:        true,
		push:           true,
		reinstallLocal: true,
		dryRun:         os.Getenv(envDashboardDryRun) == "1",
	}
	return r
}

// applyDashboardToggles mutates recipe from WRK_DASHBOARD_TOGGLES.
// Disabled gates (e.g. Add changes when clean) cannot be forced on.
func applyDashboardToggles(r *dashboardRecipe, workDir string) {
	raw := strings.TrimSpace(os.Getenv(envDashboardToggles))
	if raw == "" {
		return
	}
	addable := dashboardHasAddableDirt(workDir)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		id := strings.TrimSpace(strings.ToLower(kv[0]))
		val := strings.TrimSpace(strings.ToLower(kv[1]))
		on := val == "on" || val == "1" || val == "true"
		off := val == "off" || val == "0" || val == "false"
		if !on && !off {
			continue
		}
		switch id {
		case "add-changes", "add_changes", "addall", "add-all":
			if on {
				// Disabled gate cannot force on when clean.
				if addable {
					r.addAll = true
				}
				// else ignore force-on
			} else {
				r.addAll = false
			}
		case "gen-commit-msg", "gen_commit_msg":
			r.genCommitMsg = on
		case "commit":
			r.commit = on
		case "done":
			if on {
				r.done = true
				r.mergeBack = false
			} else if r.done {
				r.done = false
			}
		case "merge-back", "merge_back":
			if on {
				r.mergeBack = true
				r.done = false
			} else {
				r.mergeBack = false
			}
		case "sync":
			r.sync = on
		case "tag-next", "tag_next":
			r.tagNext = on
		case "push":
			r.push = on
		case "reinstall-local", "reinstall_local":
			r.reinstallLocal = on
		}
	}
}

// composeArgvFromRecipe builds CLI tokens for re-entry into wrkcli.Run.
func composeArgvFromRecipe(r dashboardRecipe) []string {
	var args []string
	if r.genCommitMsg {
		args = append(args, "--gen-commit-msg")
		if r.addAll {
			args = append(args, "--add-all")
		}
		if r.commit {
			args = append(args, "--commit")
		}
		if r.agentRunner != "" {
			args = append(args, "--agent-runner="+r.agentRunner)
		}
	}
	if r.done {
		args = append(args, "--done")
	}
	if r.mergeBack {
		args = append(args, "--merge-back")
	}
	if r.sync {
		args = append(args, "--sync")
	}
	if r.tagNext {
		args = append(args, "--tag-next")
	}
	if r.push {
		args = append(args, "--push")
	}
	if r.reinstallLocal {
		args = append(args, "--reinstall-local")
	}
	if r.dryRun {
		args = append(args, "--dry-run")
	}
	// -y is only valid on the top-level lessflags path (--done / --merge-back).
	// Bare --gen-commit-msg (no compose partner) forwards remaining flags to the
	// library, which rejects -y as "unrecognized flag". Default auto-yes already
	// covers confirm prompts when a primary is present; keep -y for hermetic
	// compose with primary for explicit compat.
	if r.done || r.mergeBack {
		args = append(args, "-y")
	}
	return args
}

// writeComposeArgvLog writes one token per line when WRK_DASHBOARD_COMPOSE_ARGV_LOG is set.
func writeComposeArgvLog(args []string) error {
	path := strings.TrimSpace(os.Getenv(envDashboardComposeArgvLog))
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Prefer one token per line (tests also accept space-joined).
	body := strings.Join(args, "\n") + "\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

// runDashboardCompose builds the recipe from defaults + env toggles, logs argv, re-enters Run.
func runDashboardCompose(workDir string, ctx *invocationContext, primaryDone bool) error {
	r := defaultDashboardRecipe(workDir, primaryDone)
	applyDashboardToggles(&r, workDir)

	// Ensure primary is set after toggles.
	if primaryDone {
		r.done = true
		r.mergeBack = false
	} else {
		r.mergeBack = true
		r.done = false
	}
	return runDashboardComposeWithRecipe(workDir, ctx, r)
}

// runDashboardComposeWithRecipe logs argv and re-enters wrkcli.Run with the given recipe.
func runDashboardComposeWithRecipe(workDir string, ctx *invocationContext, r dashboardRecipe) error {
	return runDashboardComposeWithRecipeOpts(workDir, ctx, r, true)
}

// runDashboardComposeWithRecipeOpts is the compose path used by single full-recipe
// runs and by pipeline mini-stages. When writeArgv is false, the argv log is left
// untouched (pipeline writes the full-recipe log once up front).
func runDashboardComposeWithRecipeOpts(workDir string, ctx *invocationContext, r dashboardRecipe, writeArgv bool) error {
	// Honor dirt gate: never add-all when clean even if recipe says so.
	if r.addAll && !dashboardHasAddableDirt(workDir) {
		r.addAll = false
	}

	args := composeArgvFromRecipe(r)
	if writeArgv {
		if err := writeComposeArgvLog(args); err != nil {
			return fmt.Errorf("wrk: write %s: %w", envDashboardComposeArgvLog, err)
		}
	}

	// Dry-run + Add changes: stage so gen-commit dry plan can see an index and
	// withStashedStagedForDryPlan can clear dirt for --done dry-run (Remove requires clean).
	// Restored after compose so untracked returns to untracked (zero permanent mutation).
	if r.dryRun && r.addAll {
		if err := gitRunDir(workDir, "add", "-A"); err != nil {
			return fmt.Errorf("wrk: dashboard dry-run stage for --add-all: %w", err)
		}
		defer func() {
			_ = gitRunDir(workDir, "reset", "HEAD", "--", ".")
		}()
	}

	// Outer dashboard event would overwrite primary; skip and let compose record.
	if ctx != nil {
		ctx.skipEvent = true
	}

	// Re-enter CLI compose with the same process (new invocation context).
	// workDir is already the process cwd for bare wrk; Run uses Getwd.
	return Run(args)
}

// dashboardHasAddableDirt is true when there is unstaged or untracked dirt
// (Add changes row should not be forced to [-]).
func dashboardHasAddableDirt(workDir string) bool {
	out, err := gitOutputDir(workDir, "status", "--porcelain")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if len(line) < 2 {
			return true
		}
		x, y := line[0], line[1]
		if x == '?' || y == '?' {
			return true
		}
		if y != ' ' {
			return true
		}
	}
	return false
}

// dashboardIsMainCheckout reports whether workDir is the primary (non-linked) checkout.
func dashboardIsMainCheckout(workDir string) bool {
	info, err := os.Lstat(filepath.Join(workDir, ".git"))
	if err != nil {
		return true
	}
	if info.Mode().IsRegular() {
		return false
	}
	return true
}
