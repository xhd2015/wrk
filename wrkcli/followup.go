package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeFollowupCD appends a single "cd /absolute/path" line to WRK_FOLLOWUP_FILE
// when the channel is set and follow-ups are not disabled via --no-cd.
// No-op when the env is unset/empty or disabled is true.
// Prefer writeFollowupCDTo when an invocation override is available (L2 tests).
func writeFollowupCD(disabled bool, absPath string) error {
	return writeFollowupCDTo(disabled, absPath, "")
}

// writeFollowupCDTo is like writeFollowupCD but uses followupFile when non-empty
// instead of process env WRK_FOLLOWUP_FILE (parallel-safe; no os.Setenv).
func writeFollowupCDTo(disabled bool, absPath, followupFile string) error {
	if disabled {
		return nil
	}
	outPath := strings.TrimSpace(followupFile)
	if outPath == "" {
		outPath = strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE"))
	}
	if outPath == "" {
		return nil
	}
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return nil
	}
	if !filepath.IsAbs(absPath) {
		resolved, err := filepath.Abs(absPath)
		if err != nil {
			return fmt.Errorf("resolve follow-up path: %w", err)
		}
		absPath = resolved
	}
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open follow-up file: %w", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "cd %s\n", absPath); err != nil {
		return fmt.Errorf("write follow-up: %w", err)
	}
	return nil
}

// shouldWriteHomeGatedFollowup reports whether a create follow-up cd should be
// written: true only when shellCwd equals the user home directory from
// os.UserHomeDir() (exact match after Clean + EvalSymlinks when possible).
// Empty/unresolvable home or empty shell cwd → false (fail closed).
// Does not use os.Getenv("HOME") directly.
func shouldWriteHomeGatedFollowup(shellCwd string) bool {
	shellCwd = strings.TrimSpace(shellCwd)
	if shellCwd == "" {
		return false
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return false
	}
	return sameDirPath(shellCwd, home)
}

// sameDirPath reports exact directory equality after Clean and, when possible,
// EvalSymlinks. Subdirectories under home do not match home itself.
func sameDirPath(a, b string) bool {
	return normalizeDirPath(a) == normalizeDirPath(b)
}

// normalizeDirPath returns an absolute cleaned path, resolving symlinks when possible.
func normalizeDirPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	abs = filepath.Clean(abs)
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(resolved)
	}
	return abs
}

// writeFollowupCDIfCwdIsHome writes cd dest only when shellCwd is exactly the
// user home directory. Used after successful create so non-home shells are not
// yanked by auto-cd. Still respects --no-cd and unset WRK_FOLLOWUP_FILE.
// followupFile is an optional invocation override (same as writeFollowupCDTo).
func writeFollowupCDIfCwdIsHome(disabled bool, shellCwd, dest, followupFile string) error {
	if !shouldWriteHomeGatedFollowup(shellCwd) {
		return nil
	}
	return writeFollowupCDTo(disabled, dest, followupFile)
}

// shouldWriteCwdGatedFollowup reports whether a done/set-task follow-up cd should
// be written: true only when shellCwdAtStart is non-empty and no longer exists
// on the filesystem (os.Stat not-exist). Empty path or still-existing path → false.
func shouldWriteCwdGatedFollowup(shellCwdAtStart string) bool {
	shellCwdAtStart = strings.TrimSpace(shellCwdAtStart)
	if shellCwdAtStart == "" {
		return false
	}
	_, err := os.Stat(shellCwdAtStart)
	return os.IsNotExist(err)
}

// writeFollowupCDIfCwdMissing writes cd dest only when shellCwd no longer exists.
// Used after successful --done remove / --set-task move so a surviving sibling
// or main checkout is not yanked by auto-cd. Create paths must use
// writeFollowupCDIfCwdIsHome instead.
// followupFile is an optional invocation override (same as writeFollowupCDTo).
func writeFollowupCDIfCwdMissing(disabled bool, shellCwd, dest, followupFile string) error {
	if !shouldWriteCwdGatedFollowup(shellCwd) {
		return nil
	}
	return writeFollowupCDTo(disabled, dest, followupFile)
}
