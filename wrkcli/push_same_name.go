package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// sameNameRemoteSnapshot is the pre-MergeBack probe of origin/<worktree-branch>.
// Marks are ephemeral (one compose run); they are not persisted.
type sameNameRemoteSnapshot struct {
	branch       string
	remoteExists bool
	remoteTip    string
	included     bool
	localHead    string
}

// originPushRemote returns the URL or name to use for origin ls-remote/push
// probes. Prefer remote.origin.pushurl so github.com-shaped fetch URLs with a
// local bare pushurl work offline (tests and redirect remotes).
func originPushRemote(repo string) string {
	if pushURL, err := gitOutputDir(repo, "config", "--get", "remote.origin.pushurl"); err == nil {
		if u := strings.TrimSpace(pushURL); u != "" {
			return u
		}
	}
	return "origin"
}

// snapshotSameNameOriginBranch records whether origin has the same-name branch
// as the worktree HEAD, its tip SHA, and whether that tip is already in local
// history (so a later lease-force cannot drop remote-only commits).
//
// Probe failures are non-fatal: a warning is printed and remoteExists stays
// false so land + origin/main push still run.
func snapshotSameNameOriginBranch(repo string) sameNameRemoteSnapshot {
	var snap sameNameRemoteSnapshot
	branch, err := gitOutputDir(repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not resolve worktree branch for origin probe: %v\n", err)
		return snap
	}
	branch = strings.TrimSpace(branch)
	snap.branch = branch
	if branch == "" || branch == "HEAD" {
		return snap
	}

	localHead, err := gitOutputDir(repo, "rev-parse", "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not resolve worktree HEAD for origin probe: %v\n", err)
		return snap
	}
	snap.localHead = strings.TrimSpace(localHead)

	exists, tip, err := lsRemoteOriginHead(repo, branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not probe origin/%s: %v\n", branch, err)
		return snap
	}
	if !exists || tip == "" {
		return snap
	}
	snap.remoteExists = true
	snap.remoteTip = tip

	included, incErr := commitIsAncestor(repo, tip, "HEAD")
	if incErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not check whether origin/%s tip is in local branch: %v\n", branch, incErr)
		// Fail closed: do not force-update when inclusion is unknown.
		snap.included = false
		return snap
	}
	snap.included = included
	return snap
}

func lsRemoteOriginHead(repo, branch string) (exists bool, tip string, err error) {
	if branch == "" || branch == "HEAD" {
		return false, "", nil
	}
	remote := originPushRemote(repo)
	out, _, lsErr := gitOutputDirCapture(repo, "ls-remote", "--heads", remote, branch)
	if lsErr != nil {
		out2, _, err2 := gitOutputDirCapture(repo, "ls-remote", "--heads", remote, "refs/heads/"+branch)
		if err2 != nil {
			return false, "", fmt.Errorf("git ls-remote origin %s: %w", branch, lsErr)
		}
		out = out2
	}
	tip = parseLsRemoteHeadSHA(out)
	if tip == "" {
		return false, "", nil
	}
	return true, tip, nil
}

func parseLsRemoteHeadSHA(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if looksLikeSHA(fields[0]) {
			return fields[0]
		}
	}
	return ""
}

func looksLikeSHA(s string) bool {
	if len(s) < 7 {
		return false
	}
	for _, c := range s {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
}

func commitIsAncestor(repo, ancestor, descendant string) (bool, error) {
	cmd := exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", ancestor, descendant)
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type sameNameRemoteUpdateKind int

const (
	sameNameRemoteSkipNone sameNameRemoteUpdateKind = iota
	sameNameRemoteSkipMoved
	sameNameRemoteSkipGone
	sameNameRemoteSkipNotIncluded
	sameNameRemoteDoUpdate
)

func decideSameNameRemoteUpdate(snap sameNameRemoteSnapshot, currentExists bool, currentTip string) (sameNameRemoteUpdateKind, string) {
	if !snap.remoteExists || snap.branch == "" || snap.branch == "HEAD" {
		return sameNameRemoteSkipNone, ""
	}
	if !currentExists {
		return sameNameRemoteSkipGone, fmt.Sprintf("origin/%s disappeared; not updating remote branch", snap.branch)
	}
	if currentTip != snap.remoteTip {
		return sameNameRemoteSkipMoved, fmt.Sprintf(
			"origin/%s moved (was %s, now %s); not updating remote branch",
			snap.branch, shortSHA(snap.remoteTip), shortSHA(currentTip),
		)
	}
	if !snap.included {
		return sameNameRemoteSkipNotIncluded, fmt.Sprintf(
			"origin/%s tip %s is not in local branch; not force-updating (would discard remote-only commits)",
			snap.branch, shortSHA(snap.remoteTip),
		)
	}
	return sameNameRemoteDoUpdate, ""
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// maybeUpdateSameNameOriginBranch lease-updates origin/<worktree-branch> after
// a successful land when the pre-land snapshot says that ref existed, still
// points at the same tip, and that tip was already in the local branch.
// Failures are warnings only (exit 0); land + origin/main already published.
func maybeUpdateSameNameOriginBranch(mainPath, sourcePath string, result *worktree.MergeBackResult, snap sameNameRemoteSnapshot, dryRun bool) {
	if !snap.remoteExists || snap.branch == "" || snap.branch == "HEAD" {
		return
	}

	exists, currentTip, err := lsRemoteOriginHead(mainPath, snap.branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not re-read origin/%s: %v\n", snap.branch, err)
		return
	}
	kind, warn := decideSameNameRemoteUpdate(snap, exists, currentTip)
	if warn != "" {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warn)
	}
	if kind != sameNameRemoteDoUpdate {
		return
	}

	if dryRun {
		fmt.Printf("would: git push --force-with-lease origin %s\n", snap.branch)
		return
	}

	postSHA, err := resolvePostLandSHA(mainPath, sourcePath, result, snap)
	if err != nil || postSHA == "" {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not resolve post-land tip for origin/%s: %v\n", snap.branch, err)
		} else {
			fmt.Fprintf(os.Stderr, "warning: could not resolve post-land tip for origin/%s\n", snap.branch)
		}
		return
	}

	lease := fmt.Sprintf("refs/heads/%s:%s", snap.branch, snap.remoteTip)
	refspec := postSHA + ":refs/heads/" + snap.branch
	if combined, pushErr := gitCombinedRunDir(mainPath, nil, "push", "--force-with-lease="+lease, "origin", refspec); pushErr != nil {
		msg := strings.TrimSpace(string(combined))
		if msg == "" {
			msg = pushErr.Error()
		}
		fmt.Fprintf(os.Stderr, "warning: could not update origin/%s: %s\n", snap.branch, msg)
		return
	}
	fmt.Printf("pushed %s → origin/%s\n", snap.branch, snap.branch)
}

func resolvePostLandSHA(mainPath, sourcePath string, result *worktree.MergeBackResult, snap sameNameRemoteSnapshot) (string, error) {
	action := ""
	if result != nil {
		action = result.Action
	}
	switch action {
	case "merged", "rebased-and-merged":
		return revParseVerify(mainPath, "HEAD")
	case "removed":
		if snap.localHead != "" {
			return snap.localHead, nil
		}
		return revParseVerify(mainPath, "HEAD")
	default:
		// noop / dry-run / unknown: prefer still-present worktree HEAD.
		if sourcePath != "" {
			if sha, err := revParseVerify(sourcePath, "HEAD"); err == nil && sha != "" {
				return sha, nil
			}
		}
		if snap.localHead != "" {
			return snap.localHead, nil
		}
		return revParseVerify(mainPath, "HEAD")
	}
}

func revParseVerify(repo, ref string) (string, error) {
	out, err := gitOutputDir(repo, "rev-parse", "--verify", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
