package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// followupForeignRepoMaxLevels is how many parent directories of shellCwd
// --done inspects before writing auto-cd. If any existing parent resolves to
// a different git main than the land dest, auto-cd is suppressed.
const followupForeignRepoMaxLevels = 3

// writeFollowupCD appends a single "cd /absolute/path" line to WRK_FOLLOWUP_FILE
// when the channel is set and follow-ups are not disabled via --no-cd.
// No-op when the env is unset/empty or disabled is true.
func writeFollowupCD(disabled bool, absPath string) error {
	if disabled {
		return nil
	}
	outPath := strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE"))
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
func writeFollowupCDIfCwdIsHome(disabled bool, shellCwd, dest string) error {
	if !shouldWriteHomeGatedFollowup(shellCwd) {
		return nil
	}
	return writeFollowupCD(disabled, dest)
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
// Used after successful --set-task move so a surviving sibling or main checkout
// is not yanked by auto-cd. --done uses writeFollowupCDAfterDoneRemove instead.
// Create paths must use writeFollowupCDIfCwdIsHome.
func writeFollowupCDIfCwdMissing(disabled bool, shellCwd, dest string) error {
	if !shouldWriteCwdGatedFollowup(shellCwd) {
		return nil
	}
	return writeFollowupCD(disabled, dest)
}

// writeFollowupCDAfterDoneRemove writes cd dest after a successful --done remove
// only when shellCwd no longer exists and none of up to
// followupForeignRepoMaxLevels parents of shellCwd is a different git repository
// than dest. Silent no-op when suppressed (same as the cwd-missing gate).
// --force-cd bypasses this helper entirely via forceLandInDir.
func writeFollowupCDAfterDoneRemove(disabled bool, shellCwd, dest string) error {
	if !shouldWriteCwdGatedFollowup(shellCwd) {
		return nil
	}
	if hasForeignGitRepoAncestor(shellCwd, dest, followupForeignRepoMaxLevels) {
		return nil
	}
	return writeFollowupCD(disabled, dest)
}

// hasForeignGitRepoAncestor reports whether any of the first maxLevels parents
// of shellCwd is inside a git repository whose main checkout differs from dest's
// main. Missing parents are skipped (walk continues). Non-git or unresolvable
// parents are skipped (fail open for that level). dest unresolvable → false.
func hasForeignGitRepoAncestor(shellCwd, dest string, maxLevels int) bool {
	shellCwd = strings.TrimSpace(shellCwd)
	dest = strings.TrimSpace(dest)
	if shellCwd == "" || dest == "" || maxLevels <= 0 {
		return false
	}
	destMain, err := resolveMainRepoNormalized(dest)
	if err != nil || destMain == "" {
		return false
	}
	p := shellCwd
	for i := 0; i < maxLevels; i++ {
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
		if _, err := os.Stat(p); err != nil {
			continue
		}
		parentMain, err := resolveMainRepoNormalized(p)
		if err != nil || parentMain == "" {
			continue
		}
		if parentMain != destMain {
			return true
		}
	}
	return false
}

// resolveMainRepoNormalized returns the normalized main-repo path for path, or
// an error when path is not inside a git work tree / cannot be resolved.
func resolveMainRepoNormalized(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if !worktree.IsInsideWorkTree(path) {
		return "", fmt.Errorf("not a git work tree")
	}
	top, err := worktree.ShowToplevel(path)
	if err != nil {
		return "", err
	}
	main, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return "", err
	}
	return normalizeDirPath(main), nil
}
