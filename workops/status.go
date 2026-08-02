package workops

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// Status returns a structured report for the given checkout path.
func Status(checkout string) (*StatusReport, error) {
	top, err := resolveCheckoutRoot(checkout)
	if err != nil {
		return nil, err
	}
	main, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return nil, err
	}
	mainAbs, err := normalizeAbs(main)
	if err != nil {
		return nil, err
	}

	branch, err := worktree.ReadBranch(top)
	if err != nil {
		return nil, fmt.Errorf("read branch: %w", err)
	}

	headShort, err := gitOutput(top, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		// Detached or unborn HEAD: leave empty rather than fail the whole report.
		headShort = ""
	}

	return &StatusReport{
		MainPath:     mainAbs,
		CheckoutPath: top,
		Branch:       branch,
		HeadShort:    headShort,
		IsWorktree:   worktree.IsLinked(top),
	}, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
