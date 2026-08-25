package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/bash"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/detect"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"golang.org/x/term"
)

const (
	envSpaceInvokeLog = "WRK_SPACE_INVOKE_LOG"
	envSpaceFail      = "WRK_SPACE_FAIL"
	envSpaceGOOS      = "DOT_PKGS_SPACE_GOOS"

	// Hermetic test value for WRK_SPACE_FAIL: pretend CreateAndActivate hit the 16-Desktop cap.
	spaceFailMaxDesktops = "max-desktops"

	defaultAgentRunner         = "grok-tty"
	defaultAgentPromptTemplate = "/brainstorm ${task}"
	codexAgentPromptTemplate   = "$brainstorm ${task}"
)

// defaultPromptTemplate returns the runner-aware default prompt template.
// codex and codex-tty use $brainstorm (a skill reference, not a slash
// command); all other runners use /brainstorm. Both "codex" and
// "codex-tty" are checked because config.runner is stored verbatim
// (only the --agent-runner CLI flag is normalized to codex-tty).
func defaultPromptTemplate(runner string) string {
	if runner == "codex" || runner == "codex-tty" {
		return codexAgentPromptTemplate
	}
	return defaultAgentPromptTemplate
}

// defaultAgentArgs returns the default agent-run flags before --agent-runner.
// Includes --color so TTY children (grok-tty/codex-tty) force color env.
func defaultAgentArgs() []string {
	return []string{"--session-id-from-prompt", "--no-submit", "--open", "--color"}
}

// createUXFlags are one-shot create-mode CLI flags.
type createUXFlags struct {
	newWindow     bool
	noNewWindow   bool
	newTerminal   bool
	reuseTerminal bool
	smartTerminal bool
	noNewTerminal bool
	openInAgent   bool
	noOpenInAgent bool
	// here prefers parent-shell / nested-shell agent launch (implies no window/terminal).
	here bool
	// agentRunner is a one-shot create-agent override. Nil means use config/default.
	agentRunner *string
}

func (f createUXFlags) any() bool {
	return f.newWindow || f.noNewWindow ||
		f.newTerminal || f.reuseTerminal || f.smartTerminal || f.noNewTerminal ||
		f.openInAgent || f.noOpenInAgent || f.here || f.agentRunner != nil
}

func (f createUXFlags) validate() error {
	nTerm := 0
	if f.newTerminal {
		nTerm++
	}
	if f.reuseTerminal {
		nTerm++
	}
	if f.smartTerminal {
		nTerm++
	}
	if nTerm > 1 {
		return fmt.Errorf("wrk: --new-terminal, --reuse-terminal, and --smart-terminal are mutually exclusive")
	}
	if f.openInAgent && f.noOpenInAgent {
		return fmt.Errorf("wrk: --open-in-agent and --no-open-in-agent are mutually exclusive")
	}
	noWin := f.noNewWindow || f.here
	noTerm := f.noNewTerminal || f.here
	if f.newWindow && noWin {
		return fmt.Errorf("wrk: --new-window and --no-new-window are mutually exclusive")
	}
	if noTerm && (f.newTerminal || f.reuseTerminal || f.smartTerminal) {
		return fmt.Errorf("wrk: terminal mode flags and --no-new-terminal are mutually exclusive")
	}
	if f.newWindow && noTerm {
		return fmt.Errorf("wrk: --new-window requires a terminal; cannot combine with --no-new-terminal")
	}
	return nil
}

// createUXPlan is the effective create UX after config + flags merge.
type createUXPlan struct {
	window       bool
	terminalMode string // "" | "new" | "reuse" | "smart"
	agent        bool
	here         bool // prefer follow-up / nested-shell agent launch
	runner       string
	promptTmpl   string
	agentArgs    []string
}

// resolveCreateUX builds the create UX plan. When applyConfig is true (plain
// create, no <target-dir>, no --no-config), it loads config create section
// (ignoring interceptor) then applies CLI flags. When applyConfig is false
// (create-with-target-dir or --no-config), config create.* is skipped silently
// and only CLI flags apply. In both cases window on implies terminal=new when
// terminal is still off.
func resolveCreateUX(wrkHome string, flags createUXFlags, applyConfig bool) (createUXPlan, error) {
	if err := flags.validate(); err != nil {
		return createUXPlan{}, err
	}

	plan := createUXPlan{
		runner:     defaultAgentRunner,
		promptTmpl: "", // empty = not explicitly set; resolved after runner is finalized
		agentArgs:  defaultAgentArgs(),
	}

	if applyConfig {
		cfg, err := loadConfig(wrkHome)
		if err != nil {
			return createUXPlan{}, err
		}
		if cfg != nil && cfg.Create != nil {
			c := cfg.Create
			if c.Window != nil && c.Window.Mode == "new" {
				plan.window = true
			}
			if c.Terminal != nil {
				switch c.Terminal.Mode {
				case "new", "reuse", "smart":
					plan.terminalMode = c.Terminal.Mode
				}
			}
			if c.Agent != nil {
				if c.Agent.Enabled != nil && *c.Agent.Enabled {
					plan.agent = true
				}
				if c.Agent.Runner != "" {
					plan.runner = c.Agent.Runner
				}
				if c.Agent.PromptTemplate != "" {
					plan.promptTmpl = c.Agent.PromptTemplate
				}
				if len(c.Agent.Args) > 0 {
					plan.agentArgs = append([]string(nil), c.Agent.Args...)
				}
			}
		}
	}

	// Apply CLI: negatives clear, positives set.
	// --here is an alias for --no-new-window + --no-new-terminal and selects
	// follow-up / nested-shell agent placement (distinct from plain no-new-*).
	if flags.noNewWindow || flags.here {
		plan.window = false
	}
	if flags.newWindow {
		plan.window = true
	}
	if flags.noNewTerminal || flags.here {
		plan.terminalMode = ""
	}
	if flags.newTerminal {
		plan.terminalMode = "new"
	}
	if flags.reuseTerminal {
		plan.terminalMode = "reuse"
	}
	if flags.smartTerminal {
		plan.terminalMode = "smart"
	}
	if flags.noOpenInAgent {
		plan.agent = false
	}
	if flags.openInAgent {
		plan.agent = true
	}
	if flags.here {
		plan.here = true
	}
	if flags.agentRunner != nil {
		if !plan.agent {
			return createUXPlan{}, fmt.Errorf("wrk: --agent-runner requires agent launch; remove --no-open-in-agent or pass --open-in-agent")
		}
		runner, err := normalizeCreateAgentRunner(*flags.agentRunner)
		if err != nil {
			return createUXPlan{}, err
		}
		plan.runner = runner
	}

	// Window on implies terminal new when terminal is still off.
	if plan.window && plan.terminalMode == "" {
		plan.terminalMode = "new"
	}

	// Apply runner-aware default prompt template only when the user did not
	// explicitly set prompt_template in config. Runner is finalized above.
	if plan.promptTmpl == "" {
		plan.promptTmpl = defaultPromptTemplate(plan.runner)
	}

	return plan, nil
}

// normalizeCreateAgentRunner accepts the compact runner names used by create
// UX and resolves them to the TTY providers understood by agent-run.
func normalizeCreateAgentRunner(runner string) (string, error) {
	switch runner {
	case "codex":
		return "codex-tty", nil
	case "grok":
		return "grok-tty", nil
	case "codex-tty", "grok-tty":
		return runner, nil
	default:
		return "", fmt.Errorf("wrk: unsupported create agent runner %q (want codex, codex-tty, grok, or grok-tty)", runner)
	}
}

// ensureCreateWindow runs Mission Control Desktop create+activate when plan.window
// is set. Call this BEFORE native worktree create so a hard space failure does not
// leave an orphan worktree. On success (or soft max-Desktop capacity failure),
// plan.window is cleared so runCreateUX will not create a second Desktop.
//
// When macOS is already at the Desktop maximum, CreateAndActivate returns
// space.ErrMaxDesktops: we warn and continue on the current Desktop (best-effort).
func ensureCreateWindow(plan *createUXPlan) error {
	if plan == nil || !plan.window {
		return nil
	}
	if _, err := createAndActivateSpace(); err != nil {
		if errors.Is(err, space.ErrMaxDesktops) {
			warnMaxDesktopsFallback()
			plan.window = false
			return nil
		}
		return fmt.Errorf("wrk: window: %w", err)
	}
	plan.window = false
	return nil
}

func warnMaxDesktopsFallback() {
	warnTok := "warning:"
	if term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == "" {
		warnTok = colorize("warning:", ansiOrange)
	}
	fmt.Fprintf(os.Stderr, "%s Mission Control already at maximum Desktops (16); continuing on current Desktop\n", warnTok)
}

// runCreateUX runs terminal / agent steps after native create printed the path.
// Window (space) must already be handled via ensureCreateWindow when required.
// Order: --here agent placement | terminal (optional agent follow-up) | agent-in-process.
// noCd suppresses the follow-up cd line when --here emits into WRK_FOLLOWUP_FILE.
func runCreateUX(worktreePath, task string, plan createUXPlan, noCd bool) error {
	// Defensive: if window was not pre-run (e.g. older call path), still try.
	if plan.window {
		if _, err := createAndActivateSpace(); err != nil {
			if errors.Is(err, space.ErrMaxDesktops) {
				warnMaxDesktopsFallback()
			} else {
				return fmt.Errorf("wrk: window: %w", err)
			}
		}
	}

	// --here: no iTerm/window; prefer parent-shell follow-up, else nested shell.
	if plan.here && plan.agent {
		return runHereAgent(worktreePath, plan, task, noCd)
	}

	if plan.terminalMode != "" {
		var followUps []string
		if plan.agent {
			cmd, err := buildAgentShellCommand(worktreePath, plan, task)
			if err != nil {
				return fmt.Errorf("wrk: agent follow-up: %w", err)
			}
			followUps = []string{cmd}
		}
		mode, err := itermOpenMode(plan.terminalMode)
		if err != nil {
			return err
		}
		if err := openIterm(worktreePath, mode, followUps); err != nil {
			return fmt.Errorf("wrk: terminal: %w", err)
		}
		return nil
	}

	if plan.agent {
		return runAgentInProcess(worktreePath, plan, task)
	}
	return nil
}

// runHereAgent places agent-run in the current terminal family:
// WRK_FOLLOWUP_FILE open + up-to-date bash.sh → write cd + agent-run;
// channel open but outdated bash.sh → warn, write cd, run agent in-process;
// otherwise warn and nest at the worktree (bash startup runs agent-run;
// other shells: in-process agent then LoginInteractive).
func runHereAgent(worktreePath string, plan createUXPlan, task string, noCd bool) error {
	cmd, err := buildAgentShellCommand(worktreePath, plan, task)
	if err != nil {
		return fmt.Errorf("wrk: agent follow-up: %w", err)
	}
	if followupChannelOpen() {
		if !bashIntegrationSupportsAgentFollowup() {
			warnTok := "warning:"
			if term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == "" {
				warnTok = colorize("warning:", ansiOrange)
			}
			fmt.Fprintf(os.Stderr, "%s bash integration is outdated for --here agent follow-up; update with: wrk --bash-integration --install\n", warnTok)
			if err := writeFollowupCD(noCd, worktreePath); err != nil {
				return err
			}
			return runAgentInProcess(worktreePath, plan, task)
		}
		if err := writeFollowupCD(noCd, worktreePath); err != nil {
			return err
		}
		if err := writeFollowupLine(cmd); err != nil {
			return err
		}
		return nil
	}
	return runHereAgentFallback(worktreePath, plan, task, cmd)
}

func runHereAgentFallback(worktreePath string, plan createUXPlan, task, agentCmd string) error {
	fmt.Fprintf(os.Stderr, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
	if detect.Shell() == "bash" {
		if err := runBashLoginWithAgentStartup(worktreePath, agentCmd); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return ExitCodeError{Code: exitErr.ExitCode()}
			}
			return err
		}
		return nil
	}
	if err := runAgentInProcess(worktreePath, plan, task); err != nil {
		return err
	}
	err := interactive.LoginInteractive(worktreePath, filepath.Base(worktreePath), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}

// runBashLoginWithAgentStartup starts an interactive bash at dir whose rcfile
// runs agentCmd once before the prompt (nested-shell fallback for --here).
func runBashLoginWithAgentStartup(dir, agentCmd string) error {
	prefix := filepath.Base(dir)
	rcfile, err := bash.RcFile(prefix)
	if err != nil {
		return fmt.Errorf("prepare bash rc: %w", err)
	}
	defer os.Remove(rcfile)
	f, err := os.OpenFile(rcfile, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open bash rc: %w", err)
	}
	_, werr := fmt.Fprintf(f, "\n# wrk --here agent startup\n%s\n", agentCmd)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	if cerr != nil {
		return cerr
	}
	cmd := bash.Login(dir, rcfile, "WRK_SHELL=1")
	return cmd.Run()
}

func itermOpenMode(mode string) (iterm2.OpenMode, error) {
	switch mode {
	case "new":
		return iterm2.ModeForceNew, nil
	case "reuse":
		return iterm2.ModeReuseCurrent, nil
	case "smart":
		return iterm2.ModeSmart, nil
	default:
		return 0, fmt.Errorf("wrk: unknown terminal mode %q", mode)
	}
}

// openIterm opens iTerm2 at dir with the given mode and follow-up shell commands.
// When KOOL_ITERM2_SCRIPT_OUT is set (hermetic tests), the recorded AppleScript
// embeds follow-up command text in readable form so assertions can match raw /
// shell-safe prompts; real osascript runs still use proper AppleScript escaping.
func openIterm(dir string, mode iterm2.OpenMode, followUps []string) error {
	cfg := &iterm2.Config{
		Mode:             mode,
		FollowUpCommands: followUps,
		SafeInputIgnore:  true,
	}
	if out := os.Getenv("KOOL_ITERM2_SCRIPT_OUT"); out != "" {
		cfg.Osascript = func(script string) error {
			// Restore unescaped follow-up text for script-out inspection only.
			for _, fu := range followUps {
				if fu == "" {
					continue
				}
				esc := iterm2.EscapeCommandForAppleScript(fu)
				script = strings.Replace(script, esc, fu, 1)
			}
			return os.WriteFile(out, []byte(script), 0o644)
		}
	}
	return iterm2.OpenConfig(dir, cfg)
}

// createAndActivateSpace wraps space.CreateAndActivate with a hermetic test hook:
// when WRK_SPACE_INVOKE_LOG is set, log CreateAndActivate and skip real AX / settle.
// When WRK_SPACE_FAIL=max-desktops (with the log hook), return space.ErrMaxDesktops
// after logging so capacity soft-fail can be tested without Mission Control.
func createAndActivateSpace() (int, error) {
	if logPath := os.Getenv(envSpaceInvokeLog); logPath != "" {
		if effectiveSpaceGOOS() != "darwin" {
			return 0, space.ErrUnsupportedPlatform
		}
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return 0, fmt.Errorf("space invoke log: %w", err)
		}
		_, werr := f.WriteString("CreateAndActivate\n")
		cerr := f.Close()
		if werr != nil {
			return 0, werr
		}
		if cerr != nil {
			return 0, cerr
		}
		if os.Getenv(envSpaceFail) == spaceFailMaxDesktops {
			return 0, space.ErrMaxDesktops
		}
		return 1, nil
	}
	return space.CreateAndActivate(nil)
}

func effectiveSpaceGOOS() string {
	if v := os.Getenv(envSpaceGOOS); v != "" {
		return v
	}
	return runtime.GOOS
}

func expandAgentPrompt(tmpl, task string) string {
	if tmpl == "" {
		tmpl = defaultAgentPromptTemplate
	}
	return strings.ReplaceAll(tmpl, "${task}", task)
}

// buildAgentArgv builds: agent-run run --dir <absWorktree> <args...> --agent-runner=<runner> <prompt>
// --dir is the workspace source of truth (process cwd need not equal the worktree).
// Always ensures --color so agent-run forces TTY child color even when create.agent.args
// omits it or the parent shell has NO_COLOR/TERM=dumb.
// Long prompts (agentrunapi.PromptFileSpillMinRunes) and follow-ups that would
// exceed iTerm write-text SafeMax are delivered via --prompt-file instead of a
// positional prompt, using agentrunapi.MaybeSpillPrompt.
func buildAgentArgv(worktreePath string, plan createUXPlan, task string) ([]string, error) {
	runner := plan.runner
	if runner == "" {
		runner = defaultAgentRunner
	}
	args := plan.agentArgs
	if args == nil {
		args = defaultAgentArgs()
	}
	args = ensureAgentColorArg(args)
	absDir, err := filepath.Abs(worktreePath)
	if err != nil {
		absDir = worktreePath
	}
	prompt := expandAgentPrompt(plan.promptTmpl, task)
	argv := make([]string, 0, 4+len(args)+3)
	argv = append(argv, "agent-run", "run", "--dir", absDir)
	argv = append(argv, args...)
	argv = append(argv, "--agent-runner="+runner)

	path, spilled, err := agentrunapi.MaybeSpillPrompt(prompt, agentrunapi.PromptSpillOpts{})
	if err != nil {
		return nil, err
	}
	if !spilled {
		// Prompt may be under the rune threshold while the full iTerm write-text
		// line (flags + long --dir + quoted prompt) still exceeds SafeMax.
		probe := append(append([]string{}, argv...), prompt)
		if !applescript.CheckWriteText(shellJoinArgv(probe)).OK {
			path, spilled, err = agentrunapi.MaybeSpillPrompt(prompt, agentrunapi.PromptSpillOpts{Force: true})
			if err != nil {
				return nil, err
			}
		}
	}
	if spilled {
		argv = append(argv, "--prompt-file="+path)
	} else {
		argv = append(argv, prompt)
	}
	return argv, nil
}

// ensureAgentColorArg appends --color if missing (no duplicate).
func ensureAgentColorArg(args []string) []string {
	for _, a := range args {
		if a == "--color" || strings.HasPrefix(a, "--color=") {
			return args
		}
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args...)
	out = append(out, "--color")
	return out
}

func buildAgentShellCommand(worktreePath string, plan createUXPlan, task string) (string, error) {
	argv, err := buildAgentArgv(worktreePath, plan, task)
	if err != nil {
		return "", err
	}
	return shellJoinArgv(argv), nil
}

func shellJoinArgv(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		if isSimpleShellWord(a) {
			parts[i] = a
		} else {
			parts[i] = ShellSafeQuote(a)
		}
	}
	return strings.Join(parts, " ")
}

func isSimpleShellWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '/' || r == '.' || r == '=' || r == '+':
		default:
			return false
		}
	}
	return true
}

func runAgentInProcess(worktreePath string, plan createUXPlan, task string) error {
	// --dir on argv is the workspace source of truth; process cwd may differ from worktree.
	argv, err := buildAgentArgv(worktreePath, plan, task)
	if err != nil {
		return fmt.Errorf("wrk: agent-run: %w", err)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	// Keep create's worktree path as the sole stdout contract; agent diagnostics
	// go to stderr. Interactive agent UIs typically use stderr/TTY directly.
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 0 {
				code = 1
			}
			return ExitCodeError{Code: code}
		}
		return fmt.Errorf("wrk: agent-run: %w", err)
	}
	return nil
}
