package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// Task-like positional heuristic (forgot -t / --task).
const taskLikeMaxLenBytes = 120

// isPathLikeArg reports arguments that are never treated as task descriptions.
// Path-like: contains / or \, or starts with ~, ./, or ../.
func isPathLikeArg(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, `/\`) {
		return true
	}
	if strings.HasPrefix(s, "~") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	return false
}

// isTaskLike reports whether s looks like a task description rather than a path.
// True when non-empty after trim and any of: ASCII whitespace; len > 120; len > 255.
// False when path-like.
func isTaskLike(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if isPathLikeArg(s) {
		return false
	}
	// ASCII whitespace (space, tab, LF, CR, VT, FF).
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' {
			return true
		}
	}
	if len(s) > taskLikeMaxLenBytes {
		return true
	}
	if len(s) > nameMaxComponentBytes {
		return true
	}
	return false
}

// isExistingDirArg is true when s resolves to an existing directory against cwd.
func isExistingDirArg(s, cwd string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	abs := s
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(cwd, abs)
	}
	info, err := os.Stat(filepath.Clean(abs))
	return err == nil && info.IsDir()
}

// confirmTaskLikePromote decides whether to treat a task-like positional as --task.
// kind is "target" (second positional) or "source" (first positional).
// -y auto-accepts; TTY or WRK_TASK_LIKE_CONFIRM=1 prompts; else non-TTY error + hint.
func confirmTaskLikePromote(kind, arg string, assumeYes bool) (promote bool, err error) {
	if assumeYes {
		return true, nil
	}
	interactive := term.IsTerminal(int(os.Stdin.Fd())) || os.Getenv("WRK_TASK_LIKE_CONFIRM") == "1"
	if !interactive {
		return false, taskLikeNonTTYError(kind, arg)
	}

	colorOn := term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == ""
	warnTok := "warning:"
	if colorOn {
		warnTok = colorize("warning:", ansiOrange)
	}

	switch kind {
	case "source":
		fmt.Fprintf(os.Stderr, "wrk: %s %q looks like a task description, not a source directory\n", warnTok, arg)
		fmt.Fprintf(os.Stderr, "Treat as --task (create from current directory)? [Y/n] ")
	default:
		fmt.Fprintf(os.Stderr, "wrk: %s %q looks like a task description, not a target directory\n", warnTok, arg)
		fmt.Fprintf(os.Stderr, "Treat as --task? [Y/n] ")
	}

	line, err := readStdinLineForPrompt()
	if err != nil {
		return false, fmt.Errorf("wrk: read task-like confirmation: %w", err)
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	switch answer {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("wrk: invalid input %q (expected y/n)", strings.TrimSpace(line))
	}
}

func taskLikeNonTTYError(kind, arg string) error {
	hintArg := arg
	if len(hintArg) > 40 {
		hintArg = hintArg[:37] + "..."
	}
	// Escape single quotes for shell-ish hint display.
	quoted := strings.ReplaceAll(hintArg, "'", `'"'"'`)
	switch kind {
	case "source":
		return fmt.Errorf("wrk: argument looks like a task description, not a source directory\nhint: wrk -t '%s'", quoted)
	default:
		return fmt.Errorf("wrk: second argument looks like a task description, not a target directory\nhint: pass it with -t/--task, e.g. wrk <dir> -t '%s'", quoted)
	}
}
