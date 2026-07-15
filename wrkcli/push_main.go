package wrkcli

import (
	"fmt"
	"strings"
)

// runPushMain pushes the current branch of mainRepo to its upstream remote
// (preferred) or origin + branch name (fallback). When tags is non-empty,
// also pushes those tag refs to the same remote. dryRun prints would: lines
// only and does not push.
func runPushMain(mainRepo string, dryRun bool, tags []string) error {
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
		return err
	}

	if dryRun {
		fmt.Printf("would: git push %s %s\n", remote, branch)
		for _, tag := range tags {
			fmt.Printf("would: git push %s %s\n", remote, tag)
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

	fmt.Printf("pushed %s → %s/%s\n", branch, remote, remoteBranch)
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
