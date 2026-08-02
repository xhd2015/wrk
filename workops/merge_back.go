package workops

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// MergeBack lands a linked worktree into main without removing the worktree.
// When DryRun is true, plans only (no HEAD or worktree mutations).
// Sync is reserved for post-land sync composition; when true with DryRun it
// does not mutate either. Real Sync apply is not required for Phase 1 dry-run.
func MergeBack(ctx context.Context, opts MergeBackOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if opts.WorktreeDir == "" {
		return fmt.Errorf("workops: MergeBack requires WorktreeDir")
	}

	source, err := resolveCheckoutRoot(opts.WorktreeDir)
	if err != nil {
		return err
	}
	if !worktree.IsLinked(source) {
		return fmt.Errorf("workops: MergeBack requires a linked worktree (%s is not a linked worktree)", source)
	}

	tmpDir := ""
	if opts.WrkHome != "" {
		tmpDir = filepath.Join(opts.WrkHome, "worktrees")
	}

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: source,
		TargetPath: "",
		Remove:     false,
		DryRun:     opts.DryRun,
		TmpDir:     tmpDir,
		StashLabel: "wrk-merge-back",
		// Library path is non-interactive: auto-confirm when mutation needs it.
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return true, nil
		},
	})
	if err != nil {
		return err
	}
	if result != nil && result.Action == "aborted" {
		return nil
	}

	// Sync composition: Phase 1 dry-run leaves set Sync=false. When Sync is
	// requested under DryRun, do not mutate. Full sync apply is left for a
	// later phase / caller composition.
	_ = opts.Sync
	return nil
}
