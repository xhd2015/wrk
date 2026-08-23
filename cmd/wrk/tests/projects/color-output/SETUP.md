# Scenario

**Feature**: wrk --projects conditional ANSI coloring via --color or TTY stdout

```
# pipe stdout (non-TTY) without --color -> plain text, aligned fields
wrk --projects -> no \x1b[ sequences

# --color forces ANSI even on pipe (doctest-safe)
wrk --projects --color -> highlight attention-worthy value portions only

# --color is global; other modes ignore it today
wrk --list --color -> git worktree list unchanged (no ANSI)
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- Color assertions use `assert.Output` v2 `<ansi-color>` tags with `--color` (pipe-safe).

## Steps

- Descendants record projects and set `req.Args` to `--projects` with or without `--color`, or `--list --color` for flag no-op.

## Context

- Red (`#31`): word `dirty`, count segments with N > 0, `Remote: diverged(...)`, worktree `N dirty` when N > 0, `K error` when K > 0, broken-main `Status: error: ...` value, and per-worktree `error: ...` detail values.
- Grey (`#90`): count segments with N = 0 in dirty status lines.
- Orange (`#33`): `Remote: needs push(...)` and `Remote: needs pull(...)`.
- Green (`#32`): not used on `--projects` (`clean` and `identical` stay uncolored).
- Labels (`Dir:`, `Branch:`, etc.) stay uncolored; only value substrings are wrapped.
- `Worktrees:    ` uses four spaces after the colon (aligned with other fields).

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
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

func withProjectsColor(req *Request) {
	req.Args = []string{"--projects", "--color"}
}

func stripANSI(s string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
}

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") {
		t.Fatalf("output must not contain ANSI escapes, got:\n%s", s)
	}
}

func recordColorProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func colorProjectDirLine(t *testing.T, mainRepo string) string {
	t.Helper()
	return "Dir:          " + resolvePath(t, mainRepo)
}

func colorStatusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func colorStatusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func initColorOutputRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func setupColorBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func setupColorTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initColorOutputRepo(t, repo, subject)
	runGitIsolated(t, repo, "remote", "add", "origin", originBare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	return repo
}

func pushCommitToBareOrigin(t *testing.T, workRoot, originBare, filename, content, subject string) {
	t.Helper()
	cloneDir := filepath.Join(workRoot, "origin-push-clone")
	runGitIsolated(t, workRoot, "clone", originBare, cloneDir)
	writeFile(t, filepath.Join(cloneDir, filename), content)
	runGitIsolated(t, cloneDir, "add", filename)
	runGitIsolated(t, cloneDir, "commit", "-m", subject)
	runGitIsolated(t, cloneDir, "push", "origin", "main")
}

func colorCompareWithRemoteField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + colorRemoteBriefFromResult(result)
}

func colorRemoteBriefFromResult(result *git.CompareBranchesResult) string {
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

func colorProjectStatusBlockPlain(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string) string {
	t.Helper()
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		colorProjectDirLine(t, mainRepo),
		colorStatusBranchLine(t, mainRepo),
		colorStatusCommitLine(t, mainRepo),
		statusLine,
		remoteField,
		worktreesSummary,
	)
}

func colorProjectStatusBlockTemplate(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string) string {
	t.Helper()
	return v2StdoutTemplate(colorProjectStatusBlockPlain(t, mainRepo, statusLine, remoteField, worktreesSummary))
}

func colorFormatWorktreesSummary(total, dirty, errors, prunes int) string {
	parts := []string{
		fmt.Sprintf("%d total", total),
		fmt.Sprintf("%d dirty", dirty),
	}
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("<ansi-color red>%d error</ansi-color>", errors))
	}
	if prunes > 0 {
		parts = append(parts, fmt.Sprintf("%d prune", prunes))
	}
	return strings.Join(parts, ", ")
}

func colorGitCommandCombinedError(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git_isolated.MustOutputError(t, dir, args...)
}

func colorWorktreeStatusError(t *testing.T, wtPath string) string {
	t.Helper()
	raw := colorGitCommandCombinedError(t, wtPath, "status", "--porcelain")
	return fmt.Sprintf("git status --porcelain in %s: %s", wtPath, raw)
}

func colorWorktreeErrorDetailLine(t *testing.T, wtPath, gitErr string) string {
	t.Helper()
	return fmt.Sprintf("  %s  <ansi-color red>error: %s</ansi-color>", resolvePath(t, wtPath), gitErr)
}

func colorProjectStatusBlockWithDetailsPlain(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string, detailLines []string) string {
	t.Helper()
	lines := []string{
		colorProjectDirLine(t, mainRepo),
		colorStatusBranchLine(t, mainRepo),
		colorStatusCommitLine(t, mainRepo),
		"Status:       " + statusLine,
		remoteField,
		"Worktrees:    " + worktreesSummary,
	}
	lines = append(lines, detailLines...)
	return strings.Join(lines, "\n")
}

func colorProjectStatusBlockWithDetailsTemplate(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string, detailLines []string) string {
	t.Helper()
	return v2StdoutTemplate(colorProjectStatusBlockWithDetailsPlain(t, mainRepo, statusLine, remoteField, worktreesSummary, detailLines))
}

func colorLinkedWorktreeSummary(t *testing.T, mainRepo string) string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	clean, dirty := 0, 0
	for _, entry := range linked {
		counts, err := colorGitStatusCounts(t, entry.Path)
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

type colorPorcelainCounts struct {
	staged, changed, renamed, deleted, untracked int
}

func colorGitStatusCounts(t *testing.T, repoPath string) (colorPorcelainCounts, error) {
	t.Helper()
	out := gitOutputIsolated(t, repoPath, "status", "--porcelain")
	var counts colorPorcelainCounts
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

func addColorLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func dirtyColorWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
}

func colorDirtyStatusSegment(n int, kind string) string {
	if n <= 0 {
		return fmt.Sprintf("<ansi-color #90>%d %s</ansi-color>", n, kind)
	}
	if kind == "staged" {
		return fmt.Sprintf("<ansi-color green>%d %s</ansi-color>", n, kind)
	}
	return fmt.Sprintf("<ansi-color red>%d %s</ansi-color>", n, kind)
}

func colorFormatDirtyStatusCounts(staged, changed, renamed, deleted, untracked int) string {
	return fmt.Sprintf("<ansi-color red>dirty</ansi-color> (%s, %s, %s, %s, %s)",
		colorDirtyStatusSegment(staged, "staged"),
		colorDirtyStatusSegment(changed, "changed"),
		colorDirtyStatusSegment(renamed, "renamed"),
		colorDirtyStatusSegment(deleted, "deleted"),
		colorDirtyStatusSegment(untracked, "untracked"),
	)
}

func colorProjectsOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}

func assertColorProjectsBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := colorProjectsOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between project blocks, got:\n%s", stdout)
	}
}

func ensureColorOutputHelpersUsed() {
	_ = withProjectsColor
	_ = stripANSI
	_ = assertNoANSI
	_ = recordColorProject
	_ = colorProjectDirLine
	_ = colorStatusBranchLine
	_ = colorStatusCommitLine
	_ = initColorOutputRepo
	_ = setupColorBareOrigin
	_ = setupColorTrackedMainRepo
	_ = pushCommitToBareOrigin
	_ = colorCompareWithRemoteField
	_ = colorRemoteBriefFromResult
	_ = colorProjectStatusBlockPlain
	_ = colorProjectStatusBlockTemplate
	_ = colorFormatWorktreesSummary
	_ = colorGitCommandCombinedError
	_ = colorWorktreeStatusError
	_ = colorWorktreeErrorDetailLine
	_ = colorProjectStatusBlockWithDetailsPlain
	_ = colorProjectStatusBlockWithDetailsTemplate
	_ = colorLinkedWorktreeSummary
	_ = colorGitStatusCounts
	_ = addColorLinkedWorktree
	_ = dirtyColorWorktree
	_ = colorDirtyStatusSegment
	_ = colorFormatDirtyStatusCounts
	_ = colorProjectsOutputBlockCount
	_ = assertColorProjectsBlocksSeparated
}
```