package wrkcli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// formatSharedBranchRefuse builds the user-facing refuse message for a shared
// branch checkout. op is the operation name shown after "refuse" (e.g. "--done",
// "--merge-back", "commit").
func formatSharedBranchRefuse(se *worktree.BranchSharedError, op string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: branch '%s' is checked out in multiple worktrees:\n", se.Branch)
	for _, ent := range se.Entries {
		if worktree.IsDead(ent.Path) {
			fmt.Fprintf(&b, "  %s (missing; prune with: git -C %s worktree prune)\n", ent.Path, se.MainRepo)
		} else {
			fmt.Fprintf(&b, "  %s\n", ent.Path)
		}
	}
	fmt.Fprintf(&b, "refuse %s while a branch is shared; resolve multi-checkout (or prune dead worktrees) before retrying", op)
	return b.String()
}

// mapMergeBackSharedError rewrites *worktree.BranchSharedError into the wrk
// Error:/refuse framing for the given op. Other errors pass through unchanged.
func mapMergeBackSharedError(err error, op string) error {
	if err == nil {
		return nil
	}
	var se *worktree.BranchSharedError
	if errors.As(err, &se) {
		return errors.New(formatSharedBranchRefuse(se, op))
	}
	return err
}

// refuseCommitIfBranchShared hard-fails when workDir's current branch is checked
// out in multiple worktrees. Used before gen-commit-msg --commit (including
// dry-run: fail closed). Detached HEAD is allowed.
func refuseCommitIfBranchShared(workDir string) error {
	mainRepo, err := worktree.ResolveMainRepo(workDir)
	if err != nil {
		return err
	}
	branch, err := worktree.ReadBranch(workDir)
	if err != nil {
		return err
	}
	if err := worktree.EnsureBranchExclusive(mainRepo, branch); err != nil {
		return mapMergeBackSharedError(err, "commit")
	}
	return nil
}
