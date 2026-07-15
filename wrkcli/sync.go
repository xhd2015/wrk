package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
)

// IsWipSubject reports whether a commit subject is a WIP commit for sync pass-1.
// After trimming surrounding whitespace, match is case-insensitive prefix any of:
//
//	wip:  |  wip(  |  [wip]
//
// Empty / whitespace-only subjects and mid-string-only occurrences are not WIP.
func IsWipSubject(subject string) bool {
	s := strings.ToLower(strings.TrimSpace(subject))
	if s == "" {
		return false
	}
	return strings.HasPrefix(s, "wip:") ||
		strings.HasPrefix(s, "wip(") ||
		strings.HasPrefix(s, "[wip]")
}

// syncOpts configures runSync / composition dry-run planning.
type syncOpts struct {
	DryRun bool
	// PretendMainAt, when non-empty, is used as the main tip ref for relation
	// comparisons (composition dry-run after a planned merge that has not been
	// applied yet). Real merges still use the named main branch.
	PretendMainAt string
}

// runSync performs FF-only bi-directional sync between the main checkout and
// linked named-branch worktrees.
//
// Pass 1 harvests each strictly-ahead worktree into main (merge --ff-only).
// Pass 2 distributes main into each strictly-behind worktree.
// Partial skips warn on stderr and still exit 0.
func runSync(workDir string, dryRun bool) error {
	return runSyncOpts(workDir, syncOpts{DryRun: dryRun})
}

func runSyncOpts(workDir string, opts syncOpts) error {
	dryRun := opts.DryRun
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	top, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	mainRepo, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return err
	}

	mainBranch, err := worktree.ReadBranch(mainRepo)
	if err != nil {
		return err
	}
	if mainBranch == "" || mainBranch == "HEAD" {
		return fmt.Errorf("wrk: main repository is in detached HEAD (not on a named branch)")
	}

	// mainTipRef is used for CompareBranches / WIP range checks. When composition
	// dry-run pretends the planned merge already landed, use that commit SHA.
	mainTipRef := mainBranch
	if opts.PretendMainAt != "" {
		mainTipRef = opts.PretendMainAt
	}

	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		return err
	}

	var details []string
	intoMain, intoWT, skipped := 0, 0, 0
	// Paths already warned/skipped (detached, dead, etc.) — do not double-count.
	skipOnce := map[string]struct{}{}

	// Pass 1: harvest main ← each linked named-branch worktree.
	for _, entry := range linked {
		key := syncEntryKey(entry)
		if worktree.IsDead(entry.Path) {
			if _, seen := skipOnce[key]; !seen {
				fmt.Fprintf(os.Stderr, "warning: skip %s: dead worktree\n", entry.Path)
				skipOnce[key] = struct{}{}
				skipped++
			}
			continue
		}
		if entry.Branch == "" {
			if _, seen := skipOnce[key]; !seen {
				fmt.Fprintf(os.Stderr, "warning: skip %s: detached HEAD\n", entry.Path)
				skipOnce[key] = struct{}{}
				skipped++
			}
			continue
		}

		// Refresh main tip for successive FF harvests (or pretend tip on dry-run compose).
		cmp, err := git.CompareBranches(mainRepo, mainTipRef, entry.Branch)
		if err != nil {
			return err
		}
		switch cmp.Relation {
		case git.BranchRelationSame, git.BranchRelationBIsAncestorOfA:
			// Identical or wt behind main: silent no-op for pass 1.
			continue
		case git.BranchRelationDiverged:
			if _, seen := skipOnce[key]; !seen {
				fmt.Fprintf(os.Stderr, "warning: skip %s: diverged from main\n", entry.Branch)
				skipOnce[key] = struct{}{}
				skipped++
			}
			continue
		case git.BranchRelationAIsAncestorOfB:
			// main is ancestor of wt → wt strictly ahead; candidate for harvest.
		default:
			continue
		}

		if err := worktree.IsClean(mainRepo); err != nil {
			if _, seen := skipOnce[key]; !seen {
				fmt.Fprintf(os.Stderr, "warning: skip %s: dirty main\n", entry.Branch)
				skipOnce[key] = struct{}{}
				skipped++
			}
			continue
		}

		if short, subject, ok, err := findFirstWipInRange(mainRepo, mainTipRef, entry.Branch); err != nil {
			return err
		} else if ok {
			if _, seen := skipOnce[key]; !seen {
				fmt.Fprintf(os.Stderr, "warning: skip %s: wip commit in range (%s %s)\n", entry.Branch, short, subject)
				skipOnce[key] = struct{}{}
				skipped++
			}
			continue
		}

		n := cmp.CommitsAheadB
		details = append(details, syncDetailPass1Line(entry.Branch, n))
		if !dryRun {
			if err := gitRunDir(mainRepo, "merge", "--ff-only", "--quiet", entry.Branch); err != nil {
				return fmt.Errorf("git merge --ff-only %s: %w", entry.Branch, err)
			}
			// Real harvest moves main tip; keep subsequent compares accurate.
			mainTipRef = mainBranch
		}
		intoMain++
	}

	// Pass 2: distribute each linked named-branch worktree ← main.
	for _, entry := range linked {
		key := syncEntryKey(entry)
		if _, seen := skipOnce[key]; seen {
			continue
		}
		if worktree.IsDead(entry.Path) {
			fmt.Fprintf(os.Stderr, "warning: skip %s: dead worktree\n", entry.Path)
			skipOnce[key] = struct{}{}
			skipped++
			continue
		}
		if entry.Branch == "" {
			fmt.Fprintf(os.Stderr, "warning: skip %s: detached HEAD\n", entry.Path)
			skipOnce[key] = struct{}{}
			skipped++
			continue
		}

		cmp, err := git.CompareBranches(mainRepo, mainTipRef, entry.Branch)
		if err != nil {
			return err
		}
		switch cmp.Relation {
		case git.BranchRelationSame, git.BranchRelationAIsAncestorOfB:
			// Identical or wt still ahead of main: silent no-op for pass 2.
			continue
		case git.BranchRelationDiverged:
			fmt.Fprintf(os.Stderr, "warning: skip %s: diverged from main\n", entry.Branch)
			skipOnce[key] = struct{}{}
			skipped++
			continue
		case git.BranchRelationBIsAncestorOfA:
			// wt is ancestor of main → main strictly ahead; candidate for distribute.
		default:
			continue
		}

		if err := worktree.IsClean(entry.Path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skip %s: dirty worktree\n", entry.Branch)
			skipOnce[key] = struct{}{}
			skipped++
			continue
		}

		n := cmp.CommitsAheadA
		details = append(details, syncDetailPass2Line(entry.Branch, n))
		if !dryRun {
			if err := gitRunDir(entry.Path, "merge", "--ff-only", "--quiet", mainBranch); err != nil {
				return fmt.Errorf("git merge --ff-only %s: %w", mainBranch, err)
			}
		}
		intoWT++
	}

	writeSyncStdout(details, intoMain, intoWT, skipped, dryRun)
	return nil
}

func syncEntryKey(e worktree.Entry) string {
	if e.Path != "" {
		return e.Path
	}
	return e.Branch
}

func syncCommitWord(n int) string {
	if n == 1 {
		return "commit"
	}
	return "commits"
}

func syncDetailPass1Line(branch string, n int) string {
	return fmt.Sprintf("main ← %s  (+%d %s)", branch, n, syncCommitWord(n))
}

func syncDetailPass2Line(branch string, n int) string {
	return fmt.Sprintf("%s ← main  (+%d %s)", branch, n, syncCommitWord(n))
}

func writeSyncStdout(details []string, intoMain, intoWT, skipped int, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "would: "
	}
	var b strings.Builder
	for _, d := range details {
		b.WriteString(prefix)
		b.WriteString(d)
		b.WriteByte('\n')
	}
	if len(details) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(prefix)
	fmt.Fprintf(&b, "synced: %d into main, %d into worktrees, %d skipped\n", intoMain, intoWT, skipped)
	fmt.Fprint(os.Stdout, b.String())
}

// findFirstWipInRange walks commits in mainBranch..wtBranch oldest-first and
// returns the first WIP subject's short=7 hash and subject when found.
func findFirstWipInRange(repo, mainBranch, wtBranch string) (short, subject string, found bool, err error) {
	// %H full hash, then subject; reverse = chronological (oldest first).
	out, err := gitOutputDir(repo, "log", "--reverse", "--format=%H\t%s", mainBranch+".."+wtBranch)
	if err != nil {
		return "", "", false, fmt.Errorf("git log %s..%s: %w", mainBranch, wtBranch, err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", "", false, nil
	}
	for _, line := range strings.Split(out, "\n") {
		hash, subj, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if !IsWipSubject(subj) {
			continue
		}
		shortOut, err := gitOutputDir(repo, "rev-parse", "--short=7", hash)
		if err != nil {
			// Fall back to first 7 chars of the full hash.
			if len(hash) >= 7 {
				return hash[:7], subj, true, nil
			}
			return hash, subj, true, nil
		}
		return strings.TrimSpace(shortOut), subj, true, nil
	}
	return "", "", false, nil
}
