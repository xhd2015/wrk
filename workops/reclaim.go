package workops

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
)

// ReclaimReport is the preflight for removing a linked worktree that is
// already contained in its main checkout (the GUI subset of wrk --done).
type ReclaimReport struct {
	Visible    bool
	Enabled    bool
	Reasons    []string
	SourcePath string
	TargetPath string
	Cascade    []string
}

// ReclaimOptions configures Reclaim.
type ReclaimOptions struct {
	WorktreeDir string
	WrkHome     string
}

// PlanReclaim inspects dir the same way wrk --done inspects a checkout before
// land: must be a linked worktree, porcelain-clean, and already included in
// main (same or ancestor). Nested linked worktrees under the checkout are
// cascade targets and must pass the same gates. Ahead/diverged are blockers
// here (the Code tab does not merge); wrk --done still merges those via
// MergeBack.
func PlanReclaim(ctx context.Context, dir string) (ReclaimReport, error) {
	if err := ctx.Err(); err != nil {
		return ReclaimReport{}, err
	}
	report := ReclaimReport{}
	source, err := resolveCheckoutRoot(dir)
	if err != nil {
		return report, nil
	}
	if !worktree.IsLinked(source) {
		return report, nil
	}
	report.Visible = true
	report.SourcePath = source
	mainPath, err := worktree.ReadMainRepo(source)
	if err != nil || strings.TrimSpace(mainPath) == "" {
		report.Reasons = append(report.Reasons, "Cannot resolve the main checkout for this worktree.")
		return report, nil
	}
	report.TargetPath = mainPath
	if err := inspectReclaimTarget(source, "This worktree", &report); err != nil {
		return report, err
	}
	cascade, err := ListCascadeLinkedWorktrees(ctx, source, source)
	if err != nil {
		report.Reasons = append(report.Reasons, err.Error())
		return report, nil
	}
	report.Cascade = cascade
	for _, path := range cascade {
		label := "Nested worktree " + displayReclaimPath(path)
		if err := inspectReclaimTarget(path, label, &report); err != nil {
			return report, err
		}
	}
	report.Enabled = len(report.Reasons) == 0
	return report, nil
}

// Reclaim re-runs PlanReclaim, then removes cascade targets and the source
// via worktree.MergeBack(Remove: true) — the same land function wrk --done uses.
func Reclaim(ctx context.Context, opts ReclaimOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(opts.WorktreeDir) == "" {
		return fmt.Errorf("workops: Reclaim requires WorktreeDir")
	}
	plan, err := PlanReclaim(ctx, opts.WorktreeDir)
	if err != nil {
		return err
	}
	if !plan.Visible {
		return fmt.Errorf("workops: Reclaim requires a linked worktree")
	}
	if !plan.Enabled {
		if len(plan.Reasons) == 0 {
			return fmt.Errorf("workops: worktree cannot be reclaimed")
		}
		return fmt.Errorf("%s", strings.Join(plan.Reasons, "\n"))
	}
	tmpDir := ""
	if home := strings.TrimSpace(opts.WrkHome); home != "" {
		tmpDir = filepath.Join(home, "worktrees")
	} else if def, err := resolveDefaultWrkHome(); err == nil {
		tmpDir = filepath.Join(def, "worktrees")
	}
	confirm := func(worktree.MergeBackPlan) (bool, error) { return true, nil }
	for _, path := range plan.Cascade {
		if _, err := worktree.MergeBack(worktree.MergeBackOptions{
			SourcePath: path,
			Remove:     true,
			TmpDir:     tmpDir,
			StashLabel: "wrk-merge-back",
			Confirm:    confirm,
		}); err != nil {
			return fmt.Errorf("cascade reclaim %s: %w", path, err)
		}
	}
	_, err = worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: plan.SourcePath,
		Remove:     true,
		TmpDir:     tmpDir,
		StashLabel: "wrk-merge-back",
		Confirm:    confirm,
	})
	return err
}

// ListCascadeLinkedWorktrees returns linked worktrees under consumerTop other
// than checkoutRoot. Nested independent main repos are a hard error, matching
// wrk --done D1.
func ListCascadeLinkedWorktrees(ctx context.Context, consumerTop, checkoutRoot string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result, err := scan_repo.Scan(ctx, scan_repo.Options{
		Roots:         []string{consumerTop},
		ListWorktrees: true,
		NoCache:       true,
	})
	if err != nil {
		return nil, err
	}
	cleanCheckout := filepath.Clean(checkoutRoot)
	cleanConsumer := filepath.Clean(consumerTop)
	seen := map[string]struct{}{}
	var targets []string
	addLinked := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if clean == cleanCheckout {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		if !worktree.IsLinked(path) {
			return
		}
		seen[clean] = struct{}{}
		targets = append(targets, path)
	}
	for _, repo := range result.Repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			if filepath.Clean(repo.Path) != cleanConsumer {
				return nil, fmt.Errorf("Nested main repo under this checkout blocks reclaim: %s", repo.Path)
			}
			for _, wt := range repo.Worktrees {
				if wt.IsMain {
					continue
				}
				addLinked(wt.Path)
			}
			continue
		}
		if repo.RepoType == scan_repo.RepoTypeWorktree {
			addLinked(repo.Path)
		}
		for _, wt := range repo.Worktrees {
			if wt.IsMain {
				continue
			}
			addLinked(wt.Path)
		}
	}
	return targets, nil
}

func inspectReclaimTarget(path, label string, report *ReclaimReport) error {
	if err := worktree.IsClean(path); err != nil {
		report.Reasons = append(report.Reasons, label+" has uncommitted changes.")
		return nil
	}
	mainPath, err := worktree.ReadMainRepo(path)
	if err != nil || strings.TrimSpace(mainPath) == "" {
		report.Reasons = append(report.Reasons, label+" has no main checkout.")
		return nil
	}
	branch, err := worktree.ReadBranch(path)
	if err == nil && branch != "" && branch != "HEAD" {
		if err := worktree.EnsureBranchExclusive(mainPath, branch); err != nil {
			report.Reasons = append(report.Reasons, label+": "+err.Error())
		}
	}
	head, err := git.RevParse(path, "HEAD")
	if err != nil || strings.TrimSpace(head) == "" {
		report.Reasons = append(report.Reasons, label+" has no HEAD.")
		return nil
	}
	compared, err := git.CompareBranches(mainPath, head, "HEAD")
	if err != nil {
		report.Reasons = append(report.Reasons, label+" cannot be compared to its main checkout.")
		return nil
	}
	switch compared.Relation {
	case git.BranchRelationSame, git.BranchRelationAIsAncestorOfB:
		return nil
	case git.BranchRelationBIsAncestorOfA:
		n := compared.CommitsAheadA
		if n <= 0 {
			n = 1
		}
		report.Reasons = append(report.Reasons, fmt.Sprintf("%s is %d commit%s ahead of the main checkout.", label, n, pluralCommit(n)))
	case git.BranchRelationDiverged:
		report.Reasons = append(report.Reasons, label+" has diverged from the main checkout.")
	default:
		report.Reasons = append(report.Reasons, label+" cannot be compared to its main checkout.")
	}
	return nil
}

func displayReclaimPath(path string) string {
	base := filepath.Base(filepath.Clean(path))
	parent := filepath.Base(filepath.Dir(filepath.Clean(path)))
	if parent != "" && parent != "." && parent != string(filepath.Separator) {
		return parent + "/" + base
	}
	return base
}

func pluralCommit(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
