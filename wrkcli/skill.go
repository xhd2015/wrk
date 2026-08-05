package wrkcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/skills/skillcmd"
	wrkskill "github.com/xhd2015/wrk/docs/skills/wrk"
)

const skillSubcommandName = "wrk"

// skillLocalFlags are skill-path flags that must not be treated as wrk mode flags
// when scanning skill argv (e.g. --list/-l are skill actions, not wrk --list).
var skillLocalFlags = map[string]struct{}{
	"-l":        {},
	"--list":    {},
	"--show":    {},
	"--install": {},
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
	"--propagate-tags":       {},
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
	"--open-in-agent":        {},
	"--no-open-in-agent":     {},
	"--no-config":            {},
	"--set-config":           {},
	"--main":                 {},
}

func skillUsage() string {
	return `wrk skill — embedded wrk agent skill

Usage:
  wrk skill --list|-l
  wrk skill --show [--header]
  wrk skill --install [OPTIONS] [<dir>]
  wrk skill -h|--help

Actions (exactly one):
  -l, --list       list available skills (prints "wrk")
  --show           print embedded SKILL.md (full body)
  --install        install embedded SKILL.md to agent directories

Options:
  --header         with --show: print YAML frontmatter only
  -h,--help        show this help message

Examples:
  wrk skill --list
  wrk skill --show --header
  wrk skill --install --cursor --dry-run
`
}

func wrkSingleSkill() *skillcmd.SingleSkill {
	content := wrkskill.SkillContent
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return &skillcmd.SingleSkill{
		Name:        skillSubcommandName,
		RootContent: content,
		Usage:       "wrk skill --install",
		Help:        skillUsage(),
	}
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

	return wrkSingleSkill().Handle(args)
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
