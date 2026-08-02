package workops

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// MergeBack lands a linked worktree into main without removing the worktree.
// When DryRun is true, plans only (no HEAD or worktree mutations).
// When Sync=true under DryRun, still no mutations (post-land sync is no-op).
// When Sync=true and not DryRun, runs post-land FF sync (CLI --merge-back --sync parity).
func MergeBack(ctx context.Context, opts MergeBackOptions) error {
	_, err := MergeBackFull(ctx, opts)
	return err
}

// MergeBackFull is like MergeBack but returns the land outcome for composition
// (target path, relation, action/message). Aborted confirm returns Action
// "aborted" with a nil error.
func MergeBackFull(ctx context.Context, opts MergeBackOptions) (*MergeBackResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opts.WorktreeDir == "" {
		return nil, fmt.Errorf("workops: MergeBack requires WorktreeDir")
	}

	source, err := resolveCheckoutRoot(opts.WorktreeDir)
	if err != nil {
		return nil, err
	}
	if !worktree.IsLinked(source) {
		return nil, fmt.Errorf("workops: MergeBack requires a linked worktree (%s is not a linked worktree)", source)
	}

	tmpDir := ""
	if opts.WrkHome != "" {
		tmpDir = filepath.Join(opts.WrkHome, "worktrees")
	}

	confirm := opts.Confirm
	if confirm == nil {
		// Library path is non-interactive: auto-confirm when mutation needs it.
		confirm = func(plan worktree.MergeBackPlan) (bool, error) {
			return true, nil
		}
	}

	wtResult, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: source,
		TargetPath: "",
		Remove:     false,
		DryRun:     opts.DryRun,
		TmpDir:     tmpDir,
		StashLabel: "wrk-merge-back",
		Confirm:    confirm,
	})
	if err != nil {
		return nil, err
	}

	out := toMergeBackResult(wtResult)
	if out == nil {
		return &MergeBackResult{}, nil
	}
	if out.Action == "aborted" {
		return out, nil
	}

	// Sync composition: DryRun never mutates. Non-dry-run Sync applies
	// post-land FF bi-directional sync (CLI --merge-back --sync parity).
	if opts.Sync && !opts.DryRun && out.Action != "dry-run" {
		mainPath := out.TargetPath
		if mainPath == "" {
			mainPath, err = resolveMainRepo(source)
			if err != nil {
				return out, fmt.Errorf("workops: MergeBack sync resolve main: %w", err)
			}
		}
		if err := syncLinkedWorktrees(mainPath, syncOptions{DryRun: false, Quiet: true}); err != nil {
			return out, err
		}
	}
	return out, nil
}

func toMergeBackResult(r *worktree.MergeBackResult) *MergeBackResult {
	if r == nil {
		return nil
	}
	return &MergeBackResult{
		SourcePath: r.SourcePath,
		TargetPath: r.TargetPath,
		Branch:     r.Branch,
		Relation:   r.Relation,
		Action:     r.Action,
		Message:    r.Message,
	}
}
