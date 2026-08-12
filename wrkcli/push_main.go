package wrkcli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/workops"
)

// runBarePush implements wrk --push [--force] [--dry-run]: push the current
// checkout's branch (option R — ShowToplevel of cwd, not always main).
// force uses git push --force-with-lease for the branch only.
func runBarePush(workDir string, dryRun, force bool) error {
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
	return runPushMain(checkoutRoot, dryRun, force, nil)
}

// runPushMain pushes the current branch of mainRepo to its upstream remote
// (preferred) or origin + branch name (fallback). When tags is non-empty,
// also pushes those tag refs to the same remote. dryRun prints would: lines
// only and does not push. force uses --force-with-lease for the branch only
// (tags stay non-force). Human confirmation / would: lines are printed on stdout.
//
// Core network push is workops.Push; CLI keeps would:/pushed printing.
// Confirm line stays "pushed … → …" even when force is set (not "force-pushed").
func runPushMain(mainRepo string, dryRun, force bool, tags []string) error {
	return runPushMainWithOutput(mainRepo, dryRun, force, tags, true)
}

// runPushMainWithOutput is like runPushMain. When printOutput is false (e.g.
// --tag-next --push --json), git push still runs but stdout stays clean of
// human would:/pushed lines so JSON output is not mixed with confirmations.
func runPushMainWithOutput(mainRepo string, dryRun, force bool, tags []string, printOutput bool) error {
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
				printWouldPushLines(remote, branch, tags, force)
			}
			// workops.Push DryRun also no-ops when remote missing.
			return workops.Push(context.Background(), workops.PushOptions{
				Checkout: mainRepo,
				DryRun:   true,
				Force:    force,
				Tags:     tags,
			})
		}
		return err
	}

	if dryRun {
		if printOutput {
			printWouldPushLines(remote, branch, tags, force)
		}
		return workops.Push(context.Background(), workops.PushOptions{
			Checkout: mainRepo,
			DryRun:   true,
			Force:    force,
			Tags:     tags,
		})
	}

	if err := workops.Push(context.Background(), workops.PushOptions{
		Checkout: mainRepo,
		DryRun:   false,
		Force:    force,
		Tags:     tags,
	}); err != nil {
		// Map library errors to wrk: prefix for CLI consistency where useful.
		return fmt.Errorf("wrk: %w", err)
	}

	if printOutput {
		// Stable confirm line — never "force-pushed" (D3).
		fmt.Printf("pushed %s → %s/%s\n", branch, remote, remoteBranch)
	}
	return nil
}

// printWouldPushLines emits dry-run plan lines for branch (force-with-lease when
// force) and non-force tag pushes.
func printWouldPushLines(remote, branch string, tags []string, force bool) {
	if force {
		fmt.Printf("would: git push --force-with-lease %s %s\n", remote, branch)
	} else {
		fmt.Printf("would: git push %s %s\n", remote, branch)
	}
	for _, tag := range tags {
		fmt.Printf("would: git push %s %s\n", remote, tag)
	}
}

// isNoPushRemoteErr reports whether err is the hard "no upstream and no origin"
// failure from resolvePushRemote / runPushMain. Unwind multi-main --push soft-
// skips these so pin-only consumers without origin do not fail the whole recipe.
func isNoPushRemoteErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no upstream configured") && strings.Contains(msg, "no origin remote")
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
