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

// runManualCommitStage is the wrk-owned commit path for --commit -m/--message.
// Order: refuse shared branch → optional add-all → require staged → dry-run would-line or real commit.
func runManualCommitStage(workDir, message string, noVerify, addAll, dryRun bool) error {
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
func runManualCommitPreStage(workDir string, enabled bool, message string, noVerify, addAll, dryRun bool) error {
	if !enabled {
		return nil
	}
	return runManualCommitStage(workDir, message, noVerify, addAll, dryRun)
}
