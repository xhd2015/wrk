package wrkcli

import (
	_ "embed"
	"fmt"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/skills/install"
	"github.com/xhd2015/skills/skill_file"
)

//go:embed SKILL.md
var skillContent string

const skillSubcommandName = "wrk"

// skillLocalFlags are skill-path flags that must not be treated as wrk mode flags
// when scanning skill argv (e.g. --list/-l are skill actions, not wrk --list).
var skillLocalFlags = map[string]struct{}{
	"-l":         {},
	"--list":     {},
	"--show":     {},
	"--install":  {},
	"--header":   {},
	"-h":         {},
	"--help":     {},
}

var wrkModeFlags = map[string]struct{}{
	"--done":                 {},
	"--merge-back":           {},
	"-l":                     {},
	"--list":                 {},
	"--status":               {},
	"--repos":                {},
	"--projects":             {},
	"--scan-git-repos":       {},
	"--no-cache":             {},
	"--fetch":                {},
	"--color":                {},
	"--add":                  {},
	"--rm":                   {},
	"--confirm-from-stdin":   {},
	"-y":                     {},
	"--yes":                  {},
	"--no-in-module-replace": {},
	"--all-deps":             {},
	"--tag-next":             {},
	"--sync":                 {},
	"--dep":                  {},
	"--bring":                {},
	"-t":                     {},
	"--task":                 {},
	"--set-task":             {},
	"--where":                {},
	"--cd":                   {},
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

func runSkill(origWd string, args []string, wrkHome string) error {
	if flag, found := findConflictingWrkModeFlag(args); found {
		return fmt.Errorf("wrk: %s is mutually exclusive with skill", flag)
	}

	if len(args) == 0 || isSkillHelpOnly(args) {
		fmt.Print(skillUsage())
		return nil
	}

	list, show, installAction, rest := peelSkillActionFlags(args)
	actionCount := 0
	if list {
		actionCount++
	}
	if show {
		actionCount++
	}
	if installAction {
		actionCount++
	}
	if actionCount == 0 {
		if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			return fmt.Errorf("wrk: unknown skill subcommand %q (expected --list, --show, or --install)", rest[0])
		}
		if len(rest) > 0 {
			return fmt.Errorf("wrk: unknown option %s (expected --list, --show, or --install)", rest[0])
		}
		return fmt.Errorf("wrk: skill requires exactly one of --list, --show, --install")
	}
	if actionCount > 1 {
		return fmt.Errorf("wrk: skill requires exactly one of --list, --show, --install")
	}

	switch {
	case list:
		return runSkillList(rest)
	case show:
		return runSkillShow(rest)
	default:
		return runSkillInstall(rest)
	}
}

func isSkillHelpOnly(args []string) bool {
	if len(args) == 0 {
		return true
	}
	for _, a := range args {
		if a != "-h" && a != "--help" {
			return false
		}
	}
	return true
}

// peelSkillActionFlags removes skill action flags from args, leaving action-local
// flags and positionals (e.g. --header for show, install options for install).
func peelSkillActionFlags(args []string) (list, show, installAction bool, rest []string) {
	for _, a := range args {
		switch a {
		case "-l", "--list":
			list = true
		case "--show":
			show = true
		case "--install":
			installAction = true
		default:
			rest = append(rest, a)
		}
	}
	return list, show, installAction, rest
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

func runSkillList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	fmt.Println(skillSubcommandName)
	return nil
}

func runSkillShow(args []string) error {
	headerOnly, err := parseSkillShowArgs(args)
	if err != nil {
		return err
	}
	content := skillContent
	if headerOnly {
		out, err := skill_file.FormatHeaderWithDelimiters(content)
		if err != nil {
			return fmt.Errorf("wrk: parse skill header: %w", err)
		}
		fmt.Print(out)
		return nil
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	fmt.Print(content)
	return nil
}

func parseSkillShowArgs(args []string) (headerOnly bool, err error) {
	var header bool
	remaining, err := lessflags.Bool("--header", &header).Parse(args)
	if err != nil {
		return false, err
	}
	if len(remaining) > 0 {
		return false, fmt.Errorf("wrk: unknown option %s", remaining[0])
	}
	return header, nil
}

func runSkillInstall(args []string) error {
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: skillSubcommandName,
		SkillContent: skillContent,
		Usage:        "wrk skill --install",
	}, args)
}
