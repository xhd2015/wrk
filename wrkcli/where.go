package wrkcli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

func runWhere(wrkHome, basename string) error {
	if !isBasename(basename) {
		return fmt.Errorf("wrk: --where requires a basename-only argument")
	}
	matches, err := storage.FindProjectsByBasename(wrkHome, basename)
	if err != nil {
		return err
	}
	switch len(matches) {
	case 0:
		return fmt.Errorf("wrk: no saved project for basename %q", basename)
	case 1:
		fmt.Println(matches[0])
	default:
		for _, p := range matches {
			fmt.Println(p)
		}
	}
	return nil
}

// githubPRRef is a parsed full GitHub pull request URL.
type githubPRRef struct {
	Owner  string
	Repo   string
	Number int
}

// errWherePRNeedsFullURL is the shared validation message for missing/invalid
// --where --pr refs (full GitHub PR URL only).
var errWherePRNeedsFullURL = fmt.Errorf("wrk: --where --pr requires a full GitHub pull request URL")

// parseGitHubPRURL accepts only full GitHub PR URLs:
//
//	https://github.com/<owner>/<repo>/pull/<N>
//
// Optional http://; host github.com (case-insensitive); trailing slash and
// query/fragment ignored. Bare numbers, owner/repo#N, and scheme-less forms fail.
func parseGitHubPRURL(raw string) (githubPRRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	if !strings.EqualFold(u.Hostname(), "github.com") {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")
	// Exactly owner/repo/pull/N (no extra path segments).
	if len(parts) != 4 || !strings.EqualFold(parts[2], "pull") {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	if owner == "" || repo == "" {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil || n <= 0 {
		return githubPRRef{}, errWherePRNeedsFullURL
	}
	return githubPRRef{Owner: owner, Repo: repo, Number: n}, nil
}

// runWherePR implements wrk --where --pr <full-github-pr-url>:
// parse URL → gh pr view headRefName → find local mains (projects.json + cwd)
// with matching origin → print live worktrees on that branch (lex-sorted).
func runWherePR(wrkHome, workDir, prURL string) error {
	ref, err := parseGitHubPRURL(prURL)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("Error: wrk: --pr requires the GitHub CLI (gh); install from https://cli.github.com/")
	}

	headRefName, err := ghPRHeadRefName(workDir, ref)
	if err != nil {
		return err
	}

	mains, err := findMainsMatchingOwnerRepo(wrkHome, workDir, ref.Owner, ref.Repo)
	if err != nil {
		return err
	}
	if len(mains) == 0 {
		return fmt.Errorf("wrk: no local project for %s/%s", ref.Owner, ref.Repo)
	}

	paths, err := liveWorktreePathsOnBranch(mains, headRefName)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("wrk: no local worktree for PR #%d (branch %s) in %s/%s",
			ref.Number, headRefName, ref.Owner, ref.Repo)
	}

	sort.Strings(paths)
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}

// ghPRHeadRefName runs: gh pr view N --repo owner/repo --json headRefName
// (works for OPEN/CLOSED/MERGED; any cwd is fine with --repo).
func ghPRHeadRefName(workDir string, ref githubPRRef) (string, error) {
	raw, err := runGh(workDir, "pr", "view", strconv.Itoa(ref.Number),
		"--repo", ref.Owner+"/"+ref.Repo,
		"--json", "headRefName",
	)
	if err != nil {
		return "", fmt.Errorf("wrk: gh pr view failed: %s", err.Error())
	}
	if raw == "" {
		raw = "{}"
	}
	// Decoder (not Unmarshal): hermetic fake-gh may append trailing junk.
	var view struct {
		HeadRefName string `json:"headRefName"`
	}
	dec := json.NewDecoder(strings.NewReader(raw))
	if err := dec.Decode(&view); err != nil {
		return "", fmt.Errorf("wrk: parse gh pr view JSON: %w", err)
	}
	head := strings.TrimSpace(view.HeadRefName)
	if head == "" {
		return "", fmt.Errorf("wrk: gh pr view returned empty headRefName for PR #%d", ref.Number)
	}
	return head, nil
}

// findMainsMatchingOwnerRepo returns absolute main repo paths whose origin is
// github.com matching owner/repo: all projects.json entries plus cwd's main
// when it matches (even if not yet recorded). Deduped by normalized path.
func findMainsMatchingOwnerRepo(wrkHome, workDir, owner, repo string) ([]string, error) {
	seen := make(map[string]struct{})
	var mains []string
	add := func(p string) {
		if p == "" {
			return
		}
		norm := storage.NormalizePath(p)
		if _, ok := seen[norm]; ok {
			return
		}
		if _, err := os.Stat(norm); err != nil {
			return
		}
		if !originMatchesOwnerRepo(norm, owner, repo) {
			return
		}
		seen[norm] = struct{}{}
		mains = append(mains, norm)
	}

	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		add(p)
	}

	if main, ok := storage.ResolveMainRepoForWorkDir(workDir); ok {
		add(main)
	}
	return mains, nil
}

// originMatchesOwnerRepo reports whether mainRepo's origin fetch URL is
// github.com and matches owner/repo (case-insensitive).
// Soft probe: missing origin, non-git paths, and other git errors return false
// without leaking git stderr (projects.json may list stale entries).
func originMatchesOwnerRepo(mainRepo, owner, repo string) bool {
	// Capture stderr so "No such remote 'origin'" / "not a git repository"
	// never reach the user's terminal during multi-project scan.
	originURL, _, err := gitOutputDirCapture(mainRepo, "remote", "get-url", "origin")
	if err != nil {
		return false
	}
	originURL = strings.TrimSpace(originURL)
	if !isGitHubRemoteURL(originURL) {
		return false
	}
	o, r, ok := scan_repo.ParseRemoteOwnerRepo(originURL)
	if !ok {
		return false
	}
	return strings.EqualFold(o, owner) && strings.EqualFold(r, repo)
}

// liveWorktreePathsOnBranch collects absolute paths of live (non-dead)
// worktrees across mains that have branch checked out.
func liveWorktreePathsOnBranch(mains []string, branch string) ([]string, error) {
	seen := make(map[string]struct{})
	var paths []string
	for _, main := range mains {
		entries, err := worktree.WorktreesOnBranch(main, branch)
		if err != nil {
			// Skip unreadable mains rather than aborting the whole lookup.
			continue
		}
		for _, e := range entries {
			if worktree.IsDead(e.Path) {
				continue
			}
			abs := e.Path
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(main, abs)
			}
			norm := storage.NormalizePath(abs)
			if _, ok := seen[norm]; ok {
				continue
			}
			if _, err := os.Stat(norm); err != nil {
				continue
			}
			seen[norm] = struct{}{}
			paths = append(paths, norm)
		}
	}
	return paths, nil
}
