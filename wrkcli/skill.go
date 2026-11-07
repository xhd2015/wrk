package wrkcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/skills/skillcmd"
	editdependencycode "github.com/xhd2015/wrk/docs/skills/edit-dependency-code"
	wrkskill "github.com/xhd2015/wrk/docs/skills/wrk"
	wrkpipeline "github.com/xhd2015/wrk/docs/skills/wrk-pipeline"
	wrkpipelinepublishandbump "github.com/xhd2015/wrk/docs/skills/wrk-pipeline-publish-and-bump"
)

const skillSubcommandName = "wrk"

// skillLocalFlags are skill-path flags that must not be treated as wrk mode flags
// when scanning skill argv (e.g. --list/-l are skill actions, not wrk --list).
var skillLocalFlags = map[string]struct{}{
	"-l":        {},
	"--list":    {},
	"--show":    {},
	"--install": {},
	"--version": {},
	"--header":  {},
	"-h":        {},
	"--help":    {},
}

var wrkModeFlags = map[string]struct{}{
	"--done":                 {},
	"--merge-back":           {},
	"-l":                     {},
	"--list":                 {},
	"--status":               {},
	"--repos":                {},
	"--projects":             {},
	"--projects-dep-graph":   {},
	"--scan-git-repos":       {},
	"--no-cache":             {},
	"--include-worktrees":    {},
	"--fetch":                {},
	"--github":               {},
	"--color":                {},
	"--add":                  {},
	"--rm":                   {},
	"--confirm-from-stdin":   {},
	"--confirm":              {},
	"-y":                     {},
	"--yes":                  {},
	"--no-in-module-replace": {},
	"--no-dep":               {},
	"--tag-next":             {},
	"--pr":                   {},
	"--title":                {},
	"--comment":              {},
	"--sync":                 {},
	"--bring":                {},
	"-t":                     {},
	"--task":                 {},
	"--set-task":             {},
	"--where":                {},
	"--cd":                   {},
	"--new":                  {},
	"--new-window":           {},
	"--no-new-window":        {},
	"--new-terminal":         {},
	"--reuse-terminal":       {},
	"--smart-terminal":       {},
	"--no-new-terminal":      {},
	"--here":                 {},
	"--open-in-agent":        {},
	"--no-open-in-agent":     {},
	"--agent-runner":         {},
	"--browser":              {},
	"--no-browser":           {},
	"--no-config":            {},
	"--set-config":           {},
	"--main":                 {},
	"--unwind":               {},
	"--pin-locals":           {},
	"--dep-replace":          {},
	"--dep-update":           {},
	"--undo":                 {},
}

func skillUsage() string {
	return `wrk skill — embedded agent skills for wrk workflows

Usage:
  wrk skill --list|-l
  wrk skill --show [<name>[/<topic>]] [--header]
  wrk skill <name> --show [--header]
  wrk skill --version [<name>]
  wrk skill --install [<name>[/<topic>]] [OPTIONS] [<dir>]
  wrk skill update [OPTIONS] [<dir>]
  wrk skill -h|--help

Actions (exactly one):
  -l, --list       list available skills
  --show           print a skill's SKILL.md
  --version        print metadata.version of the skill
  --install        install a skill to agent directories
  update           refresh already-installed skills
  -h,--help        show this help message

<name> defaults to "wrk" when omitted.

Skills:
  wrk                             worktrees, ship stages, PR, deps
  wrk-pipeline                    four-phase lifecycle: bring, PR, publish, cleanup
  wrk-pipeline-publish-and-bump   publish a depended-upon repo, then bump consumers
  edit-dependency-code            bring a dependency into ./external and edit it there

Install options:
  --dry-run --cursor --codex --opencode --general-agents --global
  --dir <dir> --force --no-override

Examples:
  wrk skill --list
  wrk skill --show wrk-pipeline
  wrk skill --show --header
  wrk skill --install wrk-pipeline --cursor --dry-run
  wrk skill --install --global
`
}

// wrkSkills is the single ordered list of skills hosted by the wrk binary.
// Order drives `wrk skill --list` and `wrk skill --help`.
func wrkSkills() []skillcmd.RegisteredSkill {
	return []skillcmd.RegisteredSkill{
		{Name: "wrk", Content: ensureTrailingNewline(wrkskill.SkillContent)},
		{Name: "wrk-pipeline", Content: ensureTrailingNewline(wrkpipeline.SkillContent)},
		{Name: "wrk-pipeline-publish-and-bump", Content: ensureTrailingNewline(wrkpipelinepublishandbump.SkillContent)},
		{Name: "edit-dependency-code", Content: ensureTrailingNewline(editdependencycode.SkillContent)},
	}
}

func wrkSkillRegistry() *skillcmd.Registry {
	// wrk has a single `skill` entry point (no `skills` command), so `wrk skill
	// --help` and `wrk skill` print the same multi-skill usage text.
	helpText := strings.TrimPrefix(skillUsage(), "\n")
	return &skillcmd.Registry{
		Skills:     wrkSkills(),
		Default:    skillSubcommandName,
		Usage:      "wrk skill --install",
		Help:       helpText,
		SkillsHelp: helpText,
	}
}

func ensureTrailingNewline(content string) string {
	if strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}

func runSkill(origWd string, args []string, wrkHome string) error {
	_ = origWd
	_ = wrkHome

	if flag, found := findConflictingWrkModeFlag(args); found {
		return fmt.Errorf("wrk: %s is mutually exclusive with skill", flag)
	}

	// skillcmd errors on empty argv; bare `wrk skill` prints skill-level help.
	if len(args) == 0 {
		fmt.Print(skillUsage())
		return nil
	}

	// skillcmd resolves relative install roots via filepath.Abs / os.Getwd.
	// Under Capture, pin real cwd to captureDir for the install call only so
	// dry-run/install paths match subprocess Dir=RepoDir without holding a
	// virtual-only cwd for third-party Abs.
	if captureDir != "" {
		old, err := os.Getwd()
		if err == nil {
			if err := os.Chdir(captureDir); err == nil {
				defer func() { _ = os.Chdir(old) }()
			}
		}
	}

	return wrkSkillRegistry().HandleSkills(args)
}

// findConflictingWrkModeFlag reports a wrk mode flag present in skill args.
// Skill-local action flags (-l/--list/--show/--install/...) are ignored so that
// skill --list is not treated as wrk --list.
func findConflictingWrkModeFlag(args []string) (string, bool) {
	skipValue := false
	for _, arg := range args {
		if skipValue {
			skipValue = false
			continue
		}
		if _, skill := skillLocalFlags[arg]; skill {
			continue
		}
		if _, ok := wrkModeFlags[arg]; ok {
			return arg, true
		}
		if strings.HasPrefix(arg, "-") {
			if _, ok := flagValueArgs[arg]; ok {
				skipValue = true
			}
		}
	}
	return "", false
}

// findWrkModeFlag is retained for unit tests and reports any wrk mode flag in args.
func findWrkModeFlag(args []string) (string, bool) {
	skipValue := false
	for _, arg := range args {
		if skipValue {
			skipValue = false
			continue
		}
		if _, ok := wrkModeFlags[arg]; ok {
			return arg, true
		}
		if strings.HasPrefix(arg, "-") {
			if _, ok := flagValueArgs[arg]; ok {
				skipValue = true
			}
		}
	}
	return "", false
}
