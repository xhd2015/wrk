package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

// set-config is mutually exclusive with these mode / create-flow flags.
var setConfigDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--status", "--repos", "--projects", "--projects-dep-graph",
	"--scan-git-repos", "--no-cache", "--include-worktrees",
	"--fetch", "--github", "--add", "--rm", "--where", "--cd", "--main",
	"--dep", "--bring", "--all-deps", "--tag-next", "--propagate-tags", "--sync", "--dry-run",
	"-t", "--task", "--set-task",
	"--new",
	"--exec",
	"--bash-integration", "--interceptor",
	"--no-interceptor",
}

type setConfigOpts struct {
	show   bool
	create bool

	newWindow     bool
	noNewWindow   bool
	newTerminal   bool
	reuseTerminal bool
	smartTerminal bool
	noNewTerminal bool
	openInAgent   bool
	noOpenInAgent bool

	// other disallowed tokens / positionals for mutual exclusion messages
	conflict string
}

// runSetConfig handles wrk --set-config [...].
func runSetConfig(origWd string, args []string, ctx *invocationContext) error {
	// Nested -h/--help short-circuits before parse/write (level-specific usage).
	if help, create, show := peelSetConfigHelp(args); help {
		switch {
		case create && !show:
			fmt.Print(setConfigCreateUsage())
		case show && !create:
			fmt.Print(setConfigShowUsage())
		default:
			// No action, or both create+show with help: dispatcher page.
			fmt.Print(setConfigUsage())
		}
		return nil
	}

	opts, err := parseSetConfigArgs(args)
	if err != nil {
		return err
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	ctx.wrkHome = wrkHome
	ctx.workDir = origWd
	ctx.command = "set-config"
	ctx.eventArgs = extractEventArgs(args, nil)
	if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
		return err
	}
	if err := ctx.autoRecord(); err != nil {
		return err
	}

	if opts.conflict != "" {
		return fmt.Errorf("wrk: --set-config is mutually exclusive with other modes")
	}

	if opts.show {
		if opts.create || opts.anyCreateFlag() {
			return fmt.Errorf("wrk: --set-config --show is mutually exclusive with --create flags")
		}
		return setConfigShow(wrkHome)
	}

	if !opts.create {
		return fmt.Errorf("wrk: --set-config requires --create or --show")
	}

	// Validate create UX flag conflicts under set-config.
	f := createUXFlags{
		newWindow:     opts.newWindow,
		noNewWindow:   opts.noNewWindow,
		newTerminal:   opts.newTerminal,
		reuseTerminal: opts.reuseTerminal,
		smartTerminal: opts.smartTerminal,
		noNewTerminal: opts.noNewTerminal,
		openInAgent:   opts.openInAgent,
		noOpenInAgent: opts.noOpenInAgent,
	}
	if err := f.validate(); err != nil {
		return err
	}
	if !f.any() {
		return fmt.Errorf("wrk: --set-config --create requires at least one create UX flag")
	}

	return setConfigWriteCreate(wrkHome, f)
}

// peelSetConfigHelp scans args for -h/--help and which action flags are present.
// Help may appear in any order among tokens.
func peelSetConfigHelp(args []string) (help, create, show bool) {
	for _, a := range args {
		switch a {
		case "-h", "--help":
			help = true
		case "--create":
			create = true
		case "--show":
			show = true
		}
	}
	return help, create, show
}

func setConfigUsage() string {
	return `wrk --set-config — manage $WRK_HOME/config.json

Usage:
  wrk --set-config --create [UX flags...]
  wrk --set-config --show
  wrk --set-config -h|--help

Actions (exactly one):
  --create         merge create UX defaults into config.json
  --show           pretty-print effective config.json

Options:
  -h,--help        show this help message

Run wrk --set-config --create --help for create UX flags.
Run wrk --set-config --show --help for show options.

Examples:
  wrk --set-config --create --new-window
  wrk --set-config --show
`
}

func setConfigCreateUsage() string {
	return `wrk --set-config --create — merge create UX defaults into config.json

Usage:
  wrk --set-config --create [UX flags...]
  wrk --set-config --create -h|--help

Requires at least one UX flag for a real write.
Successful write: empty stdout preferred.
Config path: $WRK_HOME/config.json

UX flags:
  --new-window           persist create.window.mode=new
                         (also sets terminal.mode=new unless a terminal flag is set)
  --new-terminal         persist create.terminal.mode=new
  --reuse-terminal       persist create.terminal.mode=reuse
  --smart-terminal       persist create.terminal.mode=smart
  --open-in-agent        enable create.agent (default runner/template/args)
  --no-new-window        clear create.window
  --no-new-terminal      clear create.terminal
  --no-open-in-agent     set create.agent.enabled=false

Conflicts (same as create mode):
  --open-in-agent with --no-open-in-agent
  --new-window with --no-new-window
  multiple terminal mode flags
  --new-window with --no-new-terminal
  terminal-on with --no-new-terminal

Notes:
  Merge-only: only keys implied by flags are written; unknown top-level keys preserved.
  --show is mutually exclusive with --create / create UX flags.

Examples:
  wrk --set-config --create --new-window
  wrk --set-config --create --open-in-agent
  wrk --set-config --create --no-open-in-agent
  wrk --set-config --create --new-window --open-in-agent
`
}

func setConfigShowUsage() string {
	return `wrk --set-config --show — print effective config.json

Usage:
  wrk --set-config --show
  wrk --set-config --show -h|--help

Prints pretty-printed JSON for $WRK_HOME/config.json.
If the file is missing, prints {} then a trailing newline.
Mutually exclusive with --create and create UX flags.

Example:
  wrk --set-config --show
`
}

func (o setConfigOpts) anyCreateFlag() bool {
	return o.newWindow || o.noNewWindow ||
		o.newTerminal || o.reuseTerminal || o.smartTerminal || o.noNewTerminal ||
		o.openInAgent || o.noOpenInAgent
}

func parseSetConfigArgs(args []string) (setConfigOpts, error) {
	var opts setConfigOpts
	sawSetConfig := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--set-config":
			if sawSetConfig {
				return opts, fmt.Errorf("wrk: duplicate --set-config")
			}
			sawSetConfig = true
		case "--show":
			opts.show = true
		case "--create":
			opts.create = true
		case "--new-window":
			opts.newWindow = true
		case "--no-new-window":
			opts.noNewWindow = true
		case "--new-terminal":
			opts.newTerminal = true
		case "--reuse-terminal":
			opts.reuseTerminal = true
		case "--smart-terminal":
			opts.smartTerminal = true
		case "--no-new-terminal":
			opts.noNewTerminal = true
		case "--open-in-agent":
			opts.openInAgent = true
		case "--no-open-in-agent":
			opts.noOpenInAgent = true
		case "--":
			// anything after is positional / conflict
			if i+1 < len(args) {
				opts.conflict = args[i+1]
			} else {
				opts.conflict = "--"
			}
			return opts, nil
		default:
			if isSetConfigDisallowed(arg) {
				opts.conflict = arg
				return opts, nil
			}
			// String flags with values that belong to other modes.
			if arg == "--add" || arg == "--rm" || arg == "--where" || arg == "--dep" || arg == "--bring" ||
				arg == "--task" || arg == "-t" || arg == "--set-task" {
				opts.conflict = arg
				return opts, nil
			}
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("wrk: unrecognized flag: %s", arg)
			}
			// Positional (e.g. create dir) is mutually exclusive.
			opts.conflict = arg
			return opts, nil
		}
	}
	if !sawSetConfig {
		return opts, fmt.Errorf("wrk: --set-config required")
	}
	return opts, nil
}

func isSetConfigDisallowed(arg string) bool {
	for _, d := range setConfigDisallowedFlags {
		if arg == d {
			return true
		}
	}
	return false
}

func setConfigShow(wrkHome string) error {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Empty object is valid JSON for show when no config yet.
			fmt.Println("{}")
			return nil
		}
		return fmt.Errorf("wrk: read config.json: %w", err)
	}
	// Pretty-print for stable readability; accept already-pretty input.
	var any interface{}
	if err := json.Unmarshal(data, &any); err != nil {
		return fmt.Errorf("wrk: parse config.json: %w", err)
	}
	out, err := json.MarshalIndent(any, "", "  ")
	if err != nil {
		return fmt.Errorf("wrk: marshal config.json: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

func setConfigWriteCreate(wrkHome string, f createUXFlags) error {
	root, err := loadConfigMap(wrkHome)
	if err != nil {
		return err
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}

	createMap, err := ensureCreateMap(root)
	if err != nil {
		return err
	}

	// Window
	if f.noNewWindow {
		delete(createMap, "window")
	}
	if f.newWindow {
		createMap["window"] = map[string]interface{}{"mode": "new"}
		// Implication: also persist terminal.mode=new unless a terminal flag sets otherwise.
		if !f.newTerminal && !f.reuseTerminal && !f.smartTerminal && !f.noNewTerminal {
			createMap["terminal"] = map[string]interface{}{"mode": "new"}
		}
	}

	// Terminal
	if f.noNewTerminal {
		delete(createMap, "terminal")
	}
	if f.newTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "new"}
	}
	if f.reuseTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "reuse"}
	}
	if f.smartTerminal {
		createMap["terminal"] = map[string]interface{}{"mode": "smart"}
	}

	// Agent
	if f.noOpenInAgent {
		agent := existingAgentMap(createMap)
		agent["enabled"] = false
		// Keep runner/template/args if present; ensure enabled is explicit.
		createMap["agent"] = agent
	}
	if f.openInAgent {
		agent := existingAgentMap(createMap)
		agent["enabled"] = true
		if _, ok := agent["runner"]; !ok || agent["runner"] == "" {
			agent["runner"] = defaultAgentRunner
		}
		if _, ok := agent["prompt_template"]; !ok || agent["prompt_template"] == "" {
			agent["prompt_template"] = defaultAgentPromptTemplate
		}
		if _, ok := agent["args"]; !ok {
			agent["args"] = defaultAgentArgs()
		}
		createMap["agent"] = agent
	}

	return saveConfigMap(wrkHome, root)
}

func existingAgentMap(createMap map[string]interface{}) map[string]interface{} {
	if v, ok := createMap["agent"]; ok && v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			// Shallow copy so we don't mutate unexpected shared maps.
			out := make(map[string]interface{}, len(m)+4)
			for k, val := range m {
				out[k] = val
			}
			return out
		}
	}
	return map[string]interface{}{}
}
