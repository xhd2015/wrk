package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

// CreateWorktreeResult is the structured outcome of default worktree creation.
type CreateWorktreeResult struct {
	Path   string
	Branch string
}

// CreateDefaultWorktree creates a linked worktree under wrkHome/worktrees using
// wrk default naming (basename-branchToken-date[-slug][-N]).
//
// projectPath may be the main repo or any checkout that resolves to it.
// wrkHome must be non-empty (callers inject Options.WrkHome or ResolveWrkHome).
// taskSlug is an already-slugified task segment; empty means no task in names.
// Callers that accept free-text tasks should use NormalizeTaskSlug first.
func CreateDefaultWorktree(projectPath, wrkHome, taskSlug string) (*CreateWorktreeResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("project path is required")
	}
	if strings.TrimSpace(wrkHome) == "" {
		return nil, fmt.Errorf("wrk home is required")
	}

	cwd, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolve project path: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return nil, fmt.Errorf("%s is not a git repository", cwd)
	}

	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return nil, err
	}

	mainRepo, err := worktree.ResolveMainRepo(checkoutRoot)
	if err != nil {
		return nil, err
	}

	baseBranch, err := worktree.ReadBranch(cwd)
	if err != nil {
		return nil, err
	}

	date := resolveWrkDate()
	_, pathToken, err := resolveNamingInputs(cwd, baseBranch)
	if err != nil {
		return nil, err
	}
	basename := filepath.Base(mainRepo)

	// Fit task slug so path/branch last components stay within 255 bytes.
	// Callers may pass already-fitted slugs; fitting is idempotent.
	fitted, fitErr := fitTaskSlugForNames(basename, pathToken, date, taskSlug)
	if fitErr != nil {
		return nil, fitErr
	}
	taskSlug = fitted

	worktreesDir := filepath.Join(wrkHome, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0o755); err != nil {
		return nil, fmt.Errorf("create worktrees dir: %w", err)
	}

	for suffix := 0; suffix < 100; suffix++ {
		wtPath, branch := candidateNames(worktreesDir, basename, pathToken, date, taskSlug, suffix)
		if candidateBlocked(mainRepo, wtPath, branch) {
			continue
		}

		if err := createWorktree(checkoutRoot, wtPath, branch); err != nil {
			return nil, err
		}

		absPath, err := filepath.Abs(wtPath)
		if err != nil {
			return nil, fmt.Errorf("resolve worktree path: %w", err)
		}
		return &CreateWorktreeResult{Path: absPath, Branch: branch}, nil
	}
	return nil, fmt.Errorf("could not find available worktree name after 99 attempts")
}

// NormalizeTaskSlug turns free-text task into a path slug.
// Empty or whitespace-only input yields empty slug (no task segment).
// Non-empty input that slugifies to empty is an error.
func NormalizeTaskSlug(taskDesc string) (string, error) {
	if strings.TrimSpace(taskDesc) == "" {
		return "", nil
	}
	slug := slugify(taskDesc)
	if slug == "" {
		return "", fmt.Errorf("task description %q produces an empty slug", taskDesc)
	}
	return slug, nil
}

// ResolveWrkHome returns wrkHome when non-empty; otherwise WRK_HOME or ~/.wrk.
func ResolveWrkHome(wrkHome string) (string, error) {
	if strings.TrimSpace(wrkHome) != "" {
		return filepath.Abs(pathfmt.Expand(wrkHome))
	}
	return resolveWrkHome()
}
