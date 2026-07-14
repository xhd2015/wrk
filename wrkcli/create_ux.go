package wrkcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	envSpaceInvokeLog = "WRK_SPACE_INVOKE_LOG"
	envSpaceGOOS      = "DOT_PKGS_SPACE_GOOS"

	defaultAgentRunner         = "grok-tty"
	defaultAgentPromptTemplate = "/brainstorm ${task}"
)

// defaultAgentArgs returns the default agent-run flags before --agent-runner.
func defaultAgentArgs() []string {
	return []string{"--session-id-from-prompt", "--no-submit", "--open"}
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
}

func (f createUXFlags) any() bool {
	return f.newWindow || f.noNewWindow ||
		f.newTerminal || f.reuseTerminal || f.smartTerminal || f.noNewTerminal ||
		f.openInAgent || f.noOpenInAgent
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
	if f.newWindow && f.noNewWindow {
		return fmt.Errorf("wrk: --new-window and --no-new-window are mutually exclusive")
	}
	if f.noNewTerminal && (f.newTerminal || f.reuseTerminal || f.smartTerminal) {
		return fmt.Errorf("wrk: terminal mode flags and --no-new-terminal are mutually exclusive")
	}
	if f.newWindow && f.noNewTerminal {
		return fmt.Errorf("wrk: --new-window requires a terminal; cannot combine with --no-new-terminal")
	}
	return nil
}

// createUXPlan is the effective create UX after config + flags merge.
type createUXPlan struct {
	window       bool
	terminalMode string // "" | "new" | "reuse" | "smart"
	agent        bool
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
		promptTmpl: defaultAgentPromptTemplate,
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
	if flags.noNewWindow {
		plan.window = false
	}
	if flags.newWindow {
		plan.window = true
	}
	if flags.noNewTerminal {
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

	// Window on implies terminal new when terminal is still off.
	if plan.window && plan.terminalMode == "" {
		plan.terminalMode = "new"
	}

	return plan, nil
}

// ensureCreateWindow runs Mission Control Desktop create+activate when plan.window
// is set. Call this BEFORE native worktree create so a space failure does not
// leave an orphan worktree. On success, plan.window is cleared so runCreateUX
// will not create a second Desktop.
func ensureCreateWindow(plan *createUXPlan) error {
	if plan == nil || !plan.window {
		return nil
	}
	if _, err := createAndActivateSpace(); err != nil {
		return fmt.Errorf("wrk: window: %w", err)
	}
	plan.window = false
	return nil
}

// runCreateUX runs terminal / agent steps after native create printed the path.
// Window (space) must already be handled via ensureCreateWindow when required.
// Order: terminal (optional agent follow-up) | agent-in-process.
func runCreateUX(worktreePath, task string, plan createUXPlan) error {
	// Defensive: if window was not pre-run (e.g. older call path), still try.
	if plan.window {
		if _, err := createAndActivateSpace(); err != nil {
			return fmt.Errorf("wrk: window: %w", err)
		}
	}

	if plan.terminalMode != "" {
		var followUps []string
		if plan.agent {
			followUps = []string{buildAgentShellCommand(worktreePath, plan, task)}
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
func buildAgentArgv(worktreePath string, plan createUXPlan, task string) []string {
	runner := plan.runner
	if runner == "" {
		runner = defaultAgentRunner
	}
	args := plan.agentArgs
	if args == nil {
		args = defaultAgentArgs()
	}
	absDir, err := filepath.Abs(worktreePath)
	if err != nil {
		absDir = worktreePath
	}
	prompt := expandAgentPrompt(plan.promptTmpl, task)
	argv := make([]string, 0, 4+len(args)+2)
	argv = append(argv, "agent-run", "run", "--dir", absDir)
	argv = append(argv, args...)
	argv = append(argv, "--agent-runner="+runner)
	argv = append(argv, prompt)
	return argv
}

func buildAgentShellCommand(worktreePath string, plan createUXPlan, task string) string {
	return shellJoinArgv(buildAgentArgv(worktreePath, plan, task))
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
	argv := buildAgentArgv(worktreePath, plan, task)
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
