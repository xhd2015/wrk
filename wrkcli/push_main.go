package wrkcli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// runBarePush implements wrk --push [--dry-run]: push the current checkout's
// branch (option R — ShowToplevel of cwd, not always main).
func runBarePush(workDir string, dryRun bool) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	return runPushMain(checkoutRoot, dryRun, nil)
}

// runPushMain pushes the current branch of mainRepo to its upstream remote
// (preferred) or origin + branch name (fallback). When tags is non-empty,
// also pushes those tag refs to the same remote. dryRun prints would: lines
// only and does not push. Human confirmation / would: lines are printed on stdout.
func runPushMain(mainRepo string, dryRun bool, tags []string) error {
	return runPushMainWithOutput(mainRepo, dryRun, tags, true)
}

// runPushMainWithOutput is like runPushMain. When printOutput is false (e.g.
// --tag-next --push --json), git push still runs but stdout stays clean of
// human would:/pushed lines so JSON output is not mixed with confirmations.
func runPushMainWithOutput(mainRepo string, dryRun bool, tags []string, printOutput bool) error {
	branch, err := gitOutputDir(mainRepo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("wrk: resolve current branch for push: %w", err)
	}
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "HEAD" {
		return fmt.Errorf("wrk: cannot push: detached HEAD on main")
	}

	remote, remoteBranch, err := resolvePushRemote(mainRepo, branch)
	if err != nil {
		// Dry-run still emits a plan when no upstream/origin is configured so
		// multi-stage compose dry-runs (e.g. dashboard RUN) can complete hermetically.
		// Real push keeps the hard error (see push/no-remote).
		if dryRun {
			remote = "origin"
			remoteBranch = branch
			if printOutput {
				fmt.Fprintf(cliStdout(), "would: git push %s %s\n", remote, branch)
				for _, tag := range tags {
					fmt.Fprintf(cliStdout(), "would: git push %s %s\n", remote, tag)
				}
			}
			return nil
		}
		return err
	}

	if dryRun {
		if printOutput {
			fmt.Fprintf(cliStdout(), "would: git push %s %s\n", remote, branch)
			for _, tag := range tags {
				fmt.Fprintf(cliStdout(), "would: git push %s %s\n", remote, tag)
			}
		}
		return nil
	}

	if out, err := gitCombinedRunDir(mainRepo, nil, "push", remote, branch); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("wrk: git push %s %s failed: %s", remote, branch, msg)
		}
		return fmt.Errorf("wrk: git push %s %s failed: %w", remote, branch, err)
	}

	for _, tag := range tags {
		if out, err := gitCombinedRunDir(mainRepo, nil, "push", remote, tag); err != nil {
			msg := strings.TrimSpace(string(out))
			if msg != "" {
				return fmt.Errorf("wrk: git push %s %s failed: %s", remote, tag, msg)
			}
			return fmt.Errorf("wrk: git push %s %s failed: %w", remote, tag, err)
		}
	}

	if printOutput {
		fmt.Fprintf(cliStdout(), "pushed %s → %s/%s\n", branch, remote, remoteBranch)
	}
	return nil
}

// resolvePushRemote returns remote and remote branch for pushing mainRepo's
// current branch. Prefer configured upstream of branch; else origin + branch.
func resolvePushRemote(mainRepo, branch string) (remote, remoteBranch string, err error) {
	remoteOut, remErr := gitOutputDir(mainRepo, "config", "--get", "branch."+branch+".remote")
	if remErr == nil {
		remote = strings.TrimSpace(remoteOut)
	}
	if remote != "" {
		mergeOut, mergeErr := gitOutputDir(mainRepo, "config", "--get", "branch."+branch+".merge")
		if mergeErr == nil {
			merge := strings.TrimSpace(mergeOut)
			remoteBranch = strings.TrimPrefix(merge, "refs/heads/")
		}
		if remoteBranch == "" {
			remoteBranch = branch
		}
		return remote, remoteBranch, nil
	}

	// Fallback: origin + current branch name (only if origin remote exists).
	if _, originErr := gitOutputDir(mainRepo, "remote", "get-url", "origin"); originErr != nil {
		return "", "", fmt.Errorf("wrk: no upstream configured for branch %q and no origin remote", branch)
	}
	return "origin", branch, nil
}
