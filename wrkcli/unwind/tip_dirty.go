package unwind

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/file"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// willUseAddAllTip is true when unwind will include unstaged/untracked tip paths
// in NextTag planning: gen-commit with --add-all --commit, or --done/--merge-back
// without gen-commit (autoCommitIfDirty scoops porcelain before land).
func willUseAddAllTip(f UnwindFlags) bool {
	addAll := f.AddAll || genArgsHasFlag(f.GenCommitArgs, "--add-all")
	if addAll && f.GenCommitMsg && genArgsHasFlag(f.GenCommitArgs, "--commit") {
		return true
	}
	return (f.Done || f.MergeBack) && !f.GenCommitMsg
}

// scopeGitPathspec maps module Dir to a git pathspec under the checkout root.
// Empty / "." means whole-repo (root scope).
func scopeGitPathspec(dir string) string {
	dir = filepath.ToSlash(strings.TrimSpace(dir))
	if dir == "" || dir == "." {
		return ""
	}
	return strings.TrimSuffix(dir, "/")
}

// tipScopeDirty reports whether owned paths under pathspec differ from releaseTag
// in a *release-relevant* way on the worktree tip.
// Mode A (addAll=false): staged only (--cached).
// Mode B (addAll=true): worktree diff + untracked (exclude-standard).
// go.sum is ignored; go.mod is compared via modfile struct (added-only replace
// does not count as dirty).
func tipScopeDirty(repoRoot, releaseTag, pathspec string, addAll bool) (bool, error) {
	repoRoot = storage.NormalizePath(repoRoot)
	if repoRoot == "" || releaseTag == "" {
		return false, nil
	}
	if addAll {
		return tipScopeDirtyAddAll(repoRoot, releaseTag, pathspec)
	}
	return tipScopeDirtyStaged(repoRoot, releaseTag, pathspec)
}

func tipScopeDirtyStaged(repoRoot, releaseTag, pathspec string) (bool, error) {
	args := []string{"diff", "--name-only", "--cached", releaseTag}
	if pathspec != "" {
		args = append(args, "--", pathspec)
	}
	out, err := gitOutputDir(repoRoot, args...)
	if err != nil {
		return false, err
	}
	return tipPathsReleaseDirty(repoRoot, releaseTag, splitNonEmptyLines(out), false)
}

func tipScopeDirtyAddAll(repoRoot, releaseTag, pathspec string) (bool, error) {
	// Worktree vs release (includes unstaged).
	args := []string{"diff", "--name-only", releaseTag}
	if pathspec != "" {
		args = append(args, "--", pathspec)
	}
	out, err := gitOutputDir(repoRoot, args...)
	if err != nil {
		return false, err
	}
	names := splitNonEmptyLines(out)

	// Index vs release (staged-only additions that match WT still need this when
	// comparing some git versions / pathspecs; keep union with WT diff).
	cargs := []string{"diff", "--name-only", "--cached", releaseTag}
	if pathspec != "" {
		cargs = append(cargs, "--", pathspec)
	}
	cout, cerr := gitOutputDir(repoRoot, cargs...)
	if cerr == nil {
		names = append(names, splitNonEmptyLines(cout)...)
	}

	uargs := []string{"ls-files", "-o", "--exclude-standard"}
	if pathspec != "" {
		uargs = append(uargs, "--", pathspec)
	}
	uout, err := gitOutputDir(repoRoot, uargs...)
	if err != nil {
		return false, err
	}
	names = append(names, splitNonEmptyLines(uout)...)
	return tipPathsReleaseDirty(repoRoot, releaseTag, names, true)
}

func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, filepath.ToSlash(line))
		}
	}
	return out
}

// tipPathsReleaseDirty classifies name-only changes: ignore go.sum; if only
// go.mod remains, structurally compare modfile vs releaseTag; else any other
// path is release-dirty.
func tipPathsReleaseDirty(repoRoot, releaseTag string, names []string, readWorktree bool) (bool, error) {
	var goModPath string
	for _, name := range names {
		base := filepath.Base(name)
		switch base {
		case "go.sum":
			continue
		case "go.mod":
			if goModPath != "" && goModPath != name {
				return true, nil
			}
			goModPath = name
		default:
			return true, nil
		}
	}
	if goModPath == "" {
		return false, nil
	}
	return goModReleaseDirty(repoRoot, releaseTag, goModPath, readWorktree)
}

// goModReleaseDirty reports whether go.mod at tip differs from releaseTag in a
// release-relevant way (structural modfile compare; added-only replace ignored).
func goModReleaseDirty(repoRoot, releaseTag, goModRel string, readWorktree bool) (bool, error) {
	beforeBlob, err := gitOutputDir(repoRoot, "show", releaseTag+":"+goModRel)
	if err != nil {
		// New go.mod vs release → release-relevant.
		return true, nil
	}
	var afterBlob string
	if readWorktree {
		data, rerr := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(goModRel)))
		if rerr != nil {
			return true, nil
		}
		afterBlob = string(data)
	} else {
		// Staged tip: index blob.
		afterBlob, err = gitOutputDir(repoRoot, "show", ":"+goModRel)
		if err != nil {
			return true, nil
		}
	}
	return goModBytesReleaseDirty([]byte(beforeBlob), []byte(afterBlob)), nil
}

// goModBytesReleaseDirty is unwind's release policy on a go.mod blob pair:
// structural Diff, ignoring added-only replace. Unparseable blobs fall back
// to trimmed raw inequality.
func goModBytesReleaseDirty(beforeData, afterData []byte) bool {
	d, err := file.DiffData(beforeData, afterData)
	if err != nil {
		return strings.TrimSpace(string(beforeData)) != strings.TrimSpace(string(afterData))
	}
	return !d.Without(file.ReplaceAdded).Empty()
}

func planRootsByLabel(members []StackMember) map[string]string {
	out := make(map[string]string, len(members))
	for label, m := range pickPeelMembersByLabel(members) {
		root := m.Path
		if root == "" {
			root = m.MainRepo
		}
		if root != "" {
			out[label] = storage.NormalizePath(root)
		}
	}
	for _, m := range members {
		if _, ok := out[m.Label]; ok {
			continue
		}
		root := m.Path
		if root == "" {
			root = m.MainRepo
		}
		if root != "" {
			out[m.Label] = storage.NormalizePath(root)
		}
	}
	return out
}

// dirtyLabels is true for stack labels with uncommitted porcelain (StackMember.Dirty).
func dirtyLabels(members []StackMember) map[string]bool {
	out := make(map[string]bool, len(members))
	for _, m := range members {
		if m.Label == "" {
			continue
		}
		if m.Dirty {
			out[m.Label] = true
		}
	}
	return out
}

// clearNextTagsOnCleanRepos drops tip-only NextTag / owned-changed on porcelain-clean
// stack members whose HEAD still sits on LatestTag. Committed releases ahead of
// LatestTag keep tagscope NextTag so verify / --tag-next still see owned-changed.
func clearNextTagsOnCleanRepos(nodes []UnwindGraphModuleNode, members []StackMember) {
	dirty := dirtyLabels(members)
	roots := planRootsByLabel(members)
	for i := range nodes {
		lab := nodes[i].RepoLabel
		if lab == "" || dirty[lab] {
			continue
		}
		if nodes[i].NextTag == "" && !nodes[i].OwnedChanged {
			continue
		}
		repo := roots[lab]
		if repo != "" && nodes[i].LatestTag != "" {
			head, herr := gitOutputDir(repo, "rev-parse", "HEAD")
			tag, terr := gitOutputDir(repo, "rev-parse", nodes[i].LatestTag+"^{commit}")
			if terr != nil {
				tag, terr = gitOutputDir(repo, "rev-parse", nodes[i].LatestTag)
			}
			head, tag = strings.TrimSpace(head), strings.TrimSpace(tag)
			if herr == nil && terr == nil && head != "" && tag != "" && head != tag {
				// HEAD ahead of latest release tag: keep tagscope next.
				continue
			}
		}
		nodes[i].NextTag = ""
		nodes[i].OwnedChanged = false
		if nodes[i].SkipReason == "" {
			nodes[i].SkipReason = "no-changes"
		}
	}
}

// refreshNextTagsFromWorktreeTip sets NextTag when HEAD-based tagscope left it
// empty but the worktree tip (staged, or staged+WT with addAll) is dirty vs LatestTag.
func refreshNextTagsFromWorktreeTip(nodes []UnwindGraphModuleNode, members []StackMember, addAll bool) {
	roots := planRootsByLabel(members)
	dirty := dirtyLabels(members)
	for i := range nodes {
		n := &nodes[i]
		if n.LatestTag == "" || n.NextTag != "" {
			continue
		}
		if n.RepoLabel == "" || !dirty[n.RepoLabel] {
			continue
		}
		repo := roots[n.RepoLabel]
		if repo == "" {
			continue
		}
		prefix := scopeGitPathspec(n.Dir)
		tipDirty, err := tipScopeDirty(repo, n.LatestTag, prefix, addAll)
		if err != nil {
			continue
		}
		// Dirty stack members with a release tag always get a next tag when the
		// tip looks release-dirty, or when HEAD still equals LatestTag (staged /
		// unpublished packages that pathspec classification may miss).
		if !tipDirty {
			// Without add-all, only staged tip dirt counts (Mode A already tried).
			if !addAll {
				continue
			}
			head, herr := gitOutputDir(repo, "rev-parse", "HEAD")
			tag, terr := gitOutputDir(repo, "rev-parse", n.LatestTag+"^{commit}")
			if terr != nil {
				tag, terr = gitOutputDir(repo, "rev-parse", n.LatestTag)
			}
			if herr != nil || terr != nil ||
				strings.TrimSpace(head) == "" ||
				strings.TrimSpace(head) != strings.TrimSpace(tag) {
				continue
			}
		}
		next, err := tagscope.IncrementTag(n.LatestTag)
		if err != nil || next == "" {
			continue
		}
		n.NextTag = next
		n.OwnedChanged = true
		n.SkipReason = ""
	}
}

// applyTipAwareTags returns a shallow snapshot copy with NextTags refreshed from
// the worktree tip and cascade rebuilt. Soft-fails to the original snap.
func applyTipAwareTags(snap *Snapshot, flags UnwindFlags) *Snapshot {
	if snap == nil || !flags.TagNext || len(snap.ModuleNodes) == 0 {
		return snap
	}
	nodes := append([]UnwindGraphModuleNode(nil), snap.ModuleNodes...)
	clearNextTagsOnCleanRepos(nodes, snap.Inv.Members)
	refreshNextTagsFromWorktreeTip(nodes, snap.Inv.Members, willUseAddAllTip(flags))
	cascade, err := planUnwindCascadeFromGraph(nodes, snap.ModuleEdges)
	if err != nil || cascade == nil {
		cascade = snap.Cascade
		if cascade == nil {
			cascade = &UnwindCascadePlan{}
		}
	}
	out := *snap
	out.ModuleNodes = nodes
	out.Cascade = cascade
	return &out
}
