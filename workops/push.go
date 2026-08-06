package workops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Push publishes the current branch (and optional tags) from checkout/main.
// DryRun is a no-op for the network (no remote ref change).
func Push(ctx context.Context, opts PushOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if opts.Checkout == "" {
		return fmt.Errorf("workops: Push requires Checkout")
	}

	checkout, err := resolveCheckoutRoot(opts.Checkout)
	if err != nil {
		return err
	}

	branch, err := gitOutput(checkout, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("workops: resolve current branch: %w", err)
	}
	if branch == "" || branch == "HEAD" {
		return fmt.Errorf("workops: cannot push: detached HEAD")
	}

	remote, remoteBranch, err := resolvePushRemote(checkout, branch)
	if err != nil {
		// Dry-run still succeeds so composition can plan hermetically without origin.
		if opts.DryRun {
			return nil
		}
		return err
	}

	if opts.DryRun {
		// Plan only — do not touch remote refs.
		_ = remote
		_ = remoteBranch
		return nil
	}

	// Branch: optional --force-with-lease (never bare --force). Tags stay non-force.
	if opts.Force {
		if err := gitRun(checkout, "push", "--force-with-lease", remote, branch); err != nil {
			return fmt.Errorf("workops: git push --force-with-lease %s %s: %w", remote, branch, err)
		}
	} else {
		if err := gitRun(checkout, "push", remote, branch); err != nil {
			return fmt.Errorf("workops: git push %s %s: %w", remote, branch, err)
		}
	}
	for _, tag := range opts.Tags {
		if tag == "" {
			continue
		}
		if err := gitRun(checkout, "push", remote, tag); err != nil {
			return fmt.Errorf("workops: git push %s %s: %w", remote, tag, err)
		}
	}
	return nil
}

func resolvePushRemote(repo, branch string) (remote, remoteBranch string, err error) {
	remoteOut, remErr := gitOutput(repo, "config", "--get", "branch."+branch+".remote")
	if remErr == nil {
		remote = strings.TrimSpace(remoteOut)
	}
	if remote != "" {
		mergeOut, mergeErr := gitOutput(repo, "config", "--get", "branch."+branch+".merge")
		if mergeErr == nil {
			merge := strings.TrimSpace(mergeOut)
			remoteBranch = strings.TrimPrefix(merge, "refs/heads/")
		}
		if remoteBranch == "" {
			remoteBranch = branch
		}
		return remote, remoteBranch, nil
	}
	if _, originErr := gitOutput(repo, "remote", "get-url", "origin"); originErr != nil {
		return "", "", fmt.Errorf("workops: no upstream for branch %q and no origin remote", branch)
	}
	return "origin", branch, nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}
