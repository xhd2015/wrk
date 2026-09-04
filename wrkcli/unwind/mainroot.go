package unwind

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// requireMainActiveRoot ensures workDir is the main repository checkout.
func requireMainActiveRoot(workDir, flag string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	if worktree.IsLinked(checkoutRoot) {
		return fmt.Errorf("wrk: %s requires the main repository checkout (activeRoot is a linked worktree, not main)", flag)
	}
	return nil
}

func gitCommandDir(repoPath string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd
}
