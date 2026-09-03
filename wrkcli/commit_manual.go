package wrkcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/git_runner"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/gitwrite"
)

// hasMessageFlag reports whether args include -m or --message (name only; value may follow).
func hasMessageFlag(args []string) bool {
	for _, a := range args {
		name := a
		if i := strings.IndexByte(a, '='); i >= 0 {
			name = a[:i]
		}
		if name == "-m" || name == "--message" {
			return true
		}
	}
	return false
}

// manualCommitShellQuote encodes msg for would:/Run: display lines (POSIX-ish).
// Mirrors agent-pro commit_msg quoting so multi-line messages stay one shell word.
func manualCommitShellQuote(s string) string {
	if !strings.ContainsAny(s, "'\"\\$ !`\n\r\t") {
		return "'" + s + "'"
	}
	return "$'" + strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	).Replace(s) + "'"
}

const noticeWorktreeCleanSkipCommit = "notice: worktree clean, skip commit\n"

// normalizeCommitMessage trims trailing whitespace so -m text matches git %B.
func normalizeCommitMessage(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}

// headCommitMessage returns HEAD's full message (%B), normalized.
func headCommitMessage(workDir string) (string, error) {
	out, err := gitOutputDir(workDir, "log", "-1", "--pretty=%B")
	if err != nil {
		return "", err
	}
	return normalizeCommitMessage(out), nil
}

// manualCommitMessageMatchesHEAD reports whether message equals HEAD's full
// commit message (trailing whitespace ignored).
func manualCommitMessageMatchesHEAD(workDir, message string) bool {
	head, err := headCommitMessage(workDir)
	if err != nil {
		return false
	}
	return normalizeCommitMessage(message) == head
}

// isNoStagedCommitErr reports empty-index failures from gen-commit-msg or
// the wrk-owned manual commit path.
func isNoStagedCommitErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no staged")
}

// runManualCommitStage is the wrk-owned commit path for --commit -m/--message.
// Order: refuse shared branch → optional add-all → require staged → dry-run would-line or real commit.
// When allowEmptySkip is true (compose partners remain) and the index is empty,
// soft-skip only if message already matches HEAD; otherwise still hard-fail.
func runManualCommitStage(workDir, message string, noVerify, addAll, dryRun, allowEmptySkip bool) error {
	if err := refuseCommitIfBranchShared(workDir); err != nil {
		return err
	}

	if addAll {
		if dryRun {
			fmt.Fprintf(os.Stderr, "would: git add -A\n")
		} else {
			fmt.Fprintf(os.Stderr, "$ git add -A\n")
			if err := gitwrite.AddAll(workDir); err != nil {
				return fmt.Errorf("git add -A failed: %w", err)
			}
		}
	}

	// Dry-run with --add-all still inspects the real index (add is plan-only),
	// matching gen-commit-msg honesty for empty staged sets.
	staged, err := git.GetStagedFiles(workDir)
	if err != nil {
		return fmt.Errorf("failed to list staged files: %w", err)
	}
	if len(staged) == 0 {
		if allowEmptySkip && manualCommitMessageMatchesHEAD(workDir, message) {
			fmt.Fprint(os.Stderr, noticeWorktreeCleanSkipCommit)
			return nil
		}
		return fmt.Errorf("no staged changes to commit")
	}

	commitCmd := "git commit -m " + manualCommitShellQuote(message)
	if noVerify {
		commitCmd += " --no-verify"
	}

	if dryRun {
		fmt.Fprintf(os.Stderr, "would: %s\n", commitCmd)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Running git commit...\n")
	output, err := git_runner.CommitWithRetry(workDir, message, 5, noVerify)
	if len(output) > 0 {
		fmt.Fprint(os.Stderr, string(output))
	}
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

// runManualCommitPreStage runs the manual commit stage before --done/--merge-back
// or as stage 1 of multi-stage compose. enabled is false when not using -m/--message.
// Always allowEmptySkip: a primary/post stage follows.
func runManualCommitPreStage(workDir string, enabled bool, message string, noVerify, addAll, dryRun bool) error {
	if !enabled {
		return nil
	}
	return runManualCommitStage(workDir, message, noVerify, addAll, dryRun, true)
}
