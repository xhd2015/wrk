package workops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// normalizeAbs returns an absolute path, resolving symlinks when possible.
func normalizeAbs(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	}
	return filepath.Clean(abs), nil
}

// resolveCheckoutRoot returns the git work-tree root (toplevel) for checkout.
func resolveCheckoutRoot(checkout string) (root string, err error) {
	abs, err := normalizeAbs(checkout)
	if err != nil {
		return "", err
	}
	if !worktree.IsInsideWorkTree(abs) {
		return "", fmt.Errorf("%s is not a git repository", abs)
	}
	top, err := worktree.ShowToplevel(abs)
	if err != nil {
		return "", fmt.Errorf("%s is not a git repository: %w", abs, err)
	}
	return normalizeAbs(top)
}

// resolveMainRepo returns the main repository absolute path for a checkout path.
func resolveMainRepo(checkout string) (mainAbs string, err error) {
	top, err := resolveCheckoutRoot(checkout)
	if err != nil {
		return "", err
	}
	main, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return "", err
	}
	return normalizeAbs(main)
}

// resolveDefaultWrkHome returns WRK_HOME or ~/.wrk when wrkHome is empty.
func resolveDefaultWrkHome() (string, error) {
	if v := strings.TrimSpace(os.Getenv("WRK_HOME")); v != "" {
		return filepath.Abs(v)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".wrk"), nil
}
