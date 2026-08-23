# Scenario

**Feature**: wrk --projects prints detailed status blocks per recorded main repo

```
# each recorded project path renders a full status block (absolute Dir)
wrk --projects -> status block per project (lexicographic order)

# extra fields vs wrk --status on main repo
Remote: <brief upstream sync summary>
Worktrees:    N total, M dirty[, K error][, P prune]  (composable segments; four spaces after colon)

# broken main repo -> minimal block (Dir + Status error only)
# per-worktree git failure -> detail line after Worktrees summary
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- `wrk --projects` is standalone; empty `projects.json` yields exit 0 and empty stdout.
- Per-project/per-worktree git failures surface inline in stdout; exit 0 unless `projects.json` is unreadable.

## Steps

- Descendants record projects via `wrk --add` or auto-record, then run `wrk --projects`.

## Context

- `Dir` is the **absolute** normalized main-repo path.
- `Remote:` uses brief sync summary from `CompareBranches(mainRepo, upstreamRef, currentBranch)`; no upstream → `(no upstream)`.
- `Worktrees:` summary uses composable segments: `N total` and `M dirty` always; `K error` when alive linked worktrees fail `git status`; `P prune` when `git worktree list` entries have missing checkout dirs (`worktree.IsDead`).
- Broken (alive, git-fails) worktrees emit a detail line: `  <abs-path>  error: <full git stderr>` (two-space indent); prunable/dead worktrees have no per-path lines.
- Broken main repo blocks omit Branch, Commit, Remote, and Worktrees; `Status: error: <full git stderr>` only.
- Blocks are separated by a blank line; project order is lexicographic by absolute path.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}

func statusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	branch := gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	return "Branch:       " + branch
}

func statusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func recordProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func projectDirLine(t *testing.T, mainRepo string) string {
	t.Helper()
	return "Dir:          " + resolvePath(t, mainRepo)
}

func formatCompareRemoteField(t *testing.T, label, upstreamRef, currentBranch string, result *git.CompareBranchesResult) string {
	t.Helper()
	body := formatKoolCompareBodyFull(upstreamRef, currentBranch, result)
	lines := strings.Split(body, "\n")
	out := label + lines[0]
	indent := strings.Repeat(" ", len(label))
	for _, line := range lines[1:] {
		out += "\n" + indent + line
	}
	return out
}

func compareWithRemoteField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + remoteBriefFromResult(result)
}

func remoteBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs push(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs pull(%d %s behind)", result.CommitsAheadA, commitWord)
	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
	default:
		return fmt.Sprintf("unknown branch relation %v", result.Relation)
	}
}

func projectStatusBlockExact(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string) string {
	t.Helper()
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		projectDirLine(t, mainRepo),
		statusBranchLine(t, mainRepo),
		statusCommitLine(t, mainRepo),
		statusLine,
		compareRemoteField,
		worktreesSummary,
	)
}

func projectStatusBlockTemplate(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string) string {
	t.Helper()
	return v2StdoutTemplate(projectStatusBlockExact(t, mainRepo, statusLine, compareRemoteField, worktreesSummary))
}

func formatWorktreesSummary(total, dirty, errors, prunes int) string {
	parts := []string{
		fmt.Sprintf("%d total", total),
		fmt.Sprintf("%d dirty", dirty),
	}
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d error", errors))
	}
	if prunes > 0 {
		parts = append(parts, fmt.Sprintf("%d prune", prunes))
	}
	return strings.Join(parts, ", ")
}

func gitCommandCombinedError(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git_isolated.MustOutputError(t, dir, args...)
}

func worktreeStatusError(t *testing.T, wtPath string) string {
	t.Helper()
	raw := gitCommandCombinedError(t, wtPath, "status", "--porcelain")
	// Product surfaces gitcmd.normalizeError shape: "git <args> in <path>: <msg>"
	return fmt.Sprintf("git status --porcelain in %s: %s", resolvePath(t, wtPath), raw)
}

func mainRepoStatusError(t *testing.T, mainRepo string) string {
	t.Helper()
	return gitCommandCombinedError(t, mainRepo, "status", "--porcelain")
}

func worktreeErrorDetailLine(t *testing.T, wtPath, gitErr string) string {
	t.Helper()
	return fmt.Sprintf("  %s  error: %s", resolvePath(t, wtPath), gitErr)
}

func projectStatusBlockWithDetailsPlain(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string, detailLines []string) string {
	t.Helper()
	lines := []string{
		projectDirLine(t, mainRepo),
		statusBranchLine(t, mainRepo),
		statusCommitLine(t, mainRepo),
		"Status:       " + statusLine,
		compareRemoteField,
		"Worktrees:    " + worktreesSummary,
	}
	lines = append(lines, detailLines...)
	return strings.Join(lines, "\n")
}

func projectStatusBlockWithDetailsTemplate(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string, detailLines []string) string {
	t.Helper()
	return v2StdoutTemplate(projectStatusBlockWithDetailsPlain(t, mainRepo, statusLine, compareRemoteField, worktreesSummary, detailLines))
}

func brokenMainRepoBlockTemplate(t *testing.T, repoPath, gitErr string) string {
	t.Helper()
	return v2StdoutTemplate(fmt.Sprintf("%s\nStatus:       error: %s",
		projectDirLine(t, repoPath), gitErr))
}

func removeGitDir(t *testing.T, repoPath string) {
	t.Helper()
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		t.Fatalf("remove %s: %v", gitDir, err)
	}
}

func removeWorktreeCheckout(t *testing.T, wtPath string) {
	t.Helper()
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("remove worktree checkout %s: %v", wtPath, err)
	}
}

func initDetailedStatusRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func setupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func setupTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initDetailedStatusRepo(t, repo, subject)
	runGitIsolated(t, repo, "remote", "add", "origin", originBare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	return repo
}

func addLinkedWorktreeForProject(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func dirtyWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
}

func linkedWorktreeSummary(t *testing.T, mainRepo string) string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	clean, dirty := 0, 0
	for _, entry := range linked {
		counts, err := gitStatusCountsForRepo(t, entry.Path)
		if err != nil {
			t.Fatalf("git status counts %q: %v", entry.Path, err)
		}
		if counts.staged == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 && counts.untracked == 0 {
			clean++
		} else {
			dirty++
		}
	}
	return fmt.Sprintf("%d total, %d dirty", clean+dirty, dirty)
}

type porcelainCounts struct {
	staged, changed, renamed, deleted, untracked int
}

func gitStatusCountsForRepo(t *testing.T, repoPath string) (porcelainCounts, error) {
	t.Helper()
	out := gitOutputIsolated(t, repoPath, "status", "--porcelain")
	var counts porcelainCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			counts.untracked++
			continue
		}
		if len(line) < 2 {
			counts.changed++
			continue
		}
		x, y := line[0], line[1]
		if x != ' ' && x != '?' {
			counts.staged++
			continue
		}
		switch {
		case y == 'R':
			counts.renamed++
		case y == 'D':
			counts.deleted++
		default:
			counts.changed++
		}
	}
	return counts, nil
}

func formatKoolCompareBodyFull(refA, refB string, result *git.CompareBranchesResult) string {
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
	default:
		return fmt.Sprintf("%s and %s diverged", refA, refB)
	}
}

func projectsOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}

func assertProjectsBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := projectsOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between project blocks, got:\n%s", stdout)
	}
}

func ensureDetailedStatusHelpersUsed() {
	_ = recordProject
	_ = projectDirLine
	_ = projectStatusBlockExact
	_ = projectStatusBlockTemplate
	_ = formatWorktreesSummary
	_ = gitCommandCombinedError
	_ = worktreeStatusError
	_ = mainRepoStatusError
	_ = worktreeErrorDetailLine
	_ = projectStatusBlockWithDetailsPlain
	_ = projectStatusBlockWithDetailsTemplate
	_ = brokenMainRepoBlockTemplate
	_ = removeGitDir
	_ = removeWorktreeCheckout
	_ = compareWithRemoteField
	_ = initDetailedStatusRepo
	_ = setupBareOrigin
	_ = setupTrackedMainRepo
	_ = addLinkedWorktreeForProject
	_ = dirtyWorktree
	_ = linkedWorktreeSummary
	_ = gitStatusCountsForRepo
	_ = formatCompareRemoteField
	_ = formatKoolCompareBodyFull
	_ = projectsOutputBlockCount
	_ = assertProjectsBlocksSeparated
}
```