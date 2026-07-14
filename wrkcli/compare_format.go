package wrkcli

import (
	"fmt"
	"strings"

	"github.com/xhd2015/gitops/git"
)

// FormatCompareBranches returns kool compare-branch text for refA vs refB.
func FormatCompareBranches(refA, refB string, result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return fmt.Sprintf("%s and %s are identical", refA, refB)

	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("%s is newer(%s +%d %s -> %s)\nto fast forward, on %s: \n   git merge --ff-only  %s",
			refB, refA, result.CommitsAheadB, commitWord, refB, refA, refB)

	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("%s is newer(%s +%d %s -> %s)\nto fast forward, on %s: \n   git merge --ff-only  %s",
			refA, refB, result.CommitsAheadA, commitWord, refA, refB, refA)

	case git.BranchRelationDiverged:
		commitWordA := "commit"
		if result.CommitsAheadA > 1 {
			commitWordA = "commits"
		}
		commitWordB := "commit"
		if result.CommitsAheadB > 1 {
			commitWordB = "commits"
		}
		return fmt.Sprintf("%s and %s has %d files difference\n"+
			"their most recent base is %s\n"+
			"%s has %d unique %s\n"+
			"%s has %d unique %s\n"+
			"They need to be merged",
			refA, refB, result.DiffFileCount, result.MergeBase,
			refA, result.CommitsAheadA, commitWordA,
			refB, result.CommitsAheadB, commitWordB)
	}

	return fmt.Sprintf("unknown branch relation %v", result.Relation)
}

// FormatMasterBrief returns a one-line master sync summary for --status linked worktrees.
func FormatMasterBrief(result *git.CompareBranchesResult, colorEnabled bool) string {
	switch result.Relation {
	case git.BranchRelationSame:
		s := "identical"
		if colorEnabled {
			return colorize(s, ansiGreen)
		}
		return s

	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("needs merge back(+%d %s)", result.CommitsAheadB, commitWord)
		if colorEnabled {
			return colorize(s, ansiOrange)
		}
		return s

	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("needs fast forward(+%d %s)", result.CommitsAheadA, commitWord)
		if colorEnabled {
			return colorize(s, ansiOrange)
		}
		return s

	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
		if colorEnabled {
			return colorize(s, ansiRed)
		}
		return s
	}

	return fmt.Sprintf("unknown branch relation %v", result.Relation)
}

// FormatRemoteBrief returns a one-line remote sync summary for --projects.
func FormatRemoteBrief(result *git.CompareBranchesResult, colorEnabled bool) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"

	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("needs push(+%d %s)", result.CommitsAheadB, commitWord)
		if colorEnabled {
			return colorize(s, ansiOrange)
		}
		return s

	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("needs pull(%d %s behind)", result.CommitsAheadA, commitWord)
		if colorEnabled {
			return colorize(s, ansiOrange)
		}
		return s

	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		s := fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
		if colorEnabled {
			return colorize(s, ansiRed)
		}
		return s
	}

	return fmt.Sprintf("unknown branch relation %v", result.Relation)
}

func formatCompareField(label, refA, refB string, result *git.CompareBranchesResult) string {
	body := FormatCompareBranches(refA, refB, result)
	lines := strings.Split(body, "\n")
	out := label + lines[0]
	indent := strings.Repeat(" ", len(label))
	for _, line := range lines[1:] {
		out += "\n" + indent + line
	}
	return out
}