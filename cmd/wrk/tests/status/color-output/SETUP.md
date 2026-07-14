# Scenario

**Feature**: wrk --status conditional ANSI coloring via --color or TTY stdout

```
# pipe stdout (non-TTY) without --color -> plain text, brief Master: labels
wrk --status -> no \x1b[ sequences

# --color forces ANSI even on pipe (doctest-safe)
wrk --status --color -> green clean, granular dirty status, colored Master: values
```

## Preconditions

- Git must be available.
- Color assertions use `assert.Output` v2 `<ansi-color>` tags with `--color` (pipe-safe).

## Steps

- Descendants set up repos/worktrees and set `req.Args` to `--status` with or without `--color`.

## Context

- Green (`#32`): entire `Status: clean` value on `--status` only; `Master: identical`.
- Red (`#31`): word `dirty`, count segments with N > 0, `Master: diverged(...)`.
- Grey (`#90`): count segments with N = 0 in dirty status lines.
- Orange (`#33`): `Master: needs merge back(...)` and `Master: needs fast forward(...)`.
- Labels (`Dir:`, `Branch:`, etc.) stay uncolored; only value substrings are wrapped.

```go
import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	return nil
}

func withStatusColor(req *Request) {
	req.Args = []string{"--status", "--color"}
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

func setupColorStatusMainRepo(t *testing.T, workRoot, name, subject string) string {
	t.Helper()
	path := filepath.Join(workRoot, name)
	statusInitRepoWithSubject(t, path, subject)
	return path
}

func addColorStatusLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func commitColorStatusOnMain(t *testing.T, mainRepo, filename, content, subject string) {
	t.Helper()
	writeFile(t, filepath.Join(mainRepo, filename), content)
	runGitIsolated(t, mainRepo, "add", filename)
	runGitIsolated(t, mainRepo, "commit", "-m", subject)
}

func commitColorStatusOnWorktree(t *testing.T, wtDir, filename, content, subject string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
	runGitIsolated(t, wtDir, "add", filename)
	runGitIsolated(t, wtDir, "commit", "-m", subject)
}

func colorStatusMasterBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs merge back(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs fast forward(+%d %s)", result.CommitsAheadA, commitWord)
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

func colorStatusMasterFieldPlain(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	return "Master:       " + colorStatusMasterBriefFromResult(result)
}

func colorStatusMasterFieldColored(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	brief := colorStatusMasterBriefFromResult(result)
	switch result.Relation {
	case git.BranchRelationSame:
		return "Master:       <ansi-color green>" + brief + "</ansi-color>"
	case git.BranchRelationAIsAncestorOfB, git.BranchRelationBIsAncestorOfA:
		return "Master:       <ansi-color #33>" + brief + "</ansi-color>"
	case git.BranchRelationDiverged:
		return "Master:       <ansi-color red>" + brief + "</ansi-color>"
	default:
		return "Master:       " + brief
	}
}

func colorStatusDirtySegment(n int, kind string) string {
	if n > 0 {
		return fmt.Sprintf("<ansi-color red>%d %s</ansi-color>", n, kind)
	}
	return fmt.Sprintf("<ansi-color #90>%d %s</ansi-color>", n, kind)
}

func colorStatusFormatDirtyCounts(added, changed, renamed, deleted int) string {
	return fmt.Sprintf("<ansi-color red>dirty</ansi-color> (%s, %s, %s, %s)",
		colorStatusDirtySegment(added, "added"),
		colorStatusDirtySegment(changed, "changed"),
		colorStatusDirtySegment(renamed, "renamed"),
		colorStatusDirtySegment(deleted, "deleted"),
	)
}

func colorStatusBlockPlain(t *testing.T, repoDir, relDir, statusLine, masterLine string) string {
	t.Helper()
	var block string
	if relDir == "." {
		block = statusRootBlockPlain(t, repoDir, statusLine, statusNoUpstreamRemote())
	} else {
		block = fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s",
			relDir, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir), statusLine)
	}
	if masterLine != "" {
		block += "\n" + masterLine
	}
	return block
}

func colorStatusBlockTemplate(t *testing.T, repoDir, relDir, statusLine, masterLine string) string {
	t.Helper()
	return v2StdoutTemplate(colorStatusBlockPlain(t, repoDir, relDir, statusLine, masterLine))
}

func colorStatusBlockContains(t *testing.T, repoDir, relDir, statusLine, masterLine string) string {
	t.Helper()
	return colorStatusBlockTemplate(t, repoDir, relDir, statusLine, masterLine)
}

func colorStatusStdoutV2(t *testing.T, blocks ...string) string {
	t.Helper()
	return v2StdoutTemplate(joinStdoutBlocks(blocks...))
}

func dirtyColorStatusRepo(t *testing.T, repo string) {
	t.Helper()
	writeFile(t, filepath.Join(repo, "added.txt"), "added\n")
	runGitIsolated(t, repo, "add", "added.txt")
	writeFile(t, filepath.Join(repo, "README.md"), "# changed\n")
}

func ensureColorStatusHelpersUsed() {
	_ = withStatusColor
	_ = stripANSI
	_ = assertNoANSI
	_ = setupColorStatusMainRepo
	_ = addColorStatusLinkedWorktree
	_ = commitColorStatusOnMain
	_ = commitColorStatusOnWorktree
	_ = colorStatusMasterBriefFromResult
	_ = colorStatusMasterFieldPlain
	_ = colorStatusMasterFieldColored
	_ = colorStatusDirtySegment
	_ = colorStatusFormatDirtyCounts
	_ = colorStatusBlockPlain
	_ = colorStatusBlockTemplate
	_ = colorStatusBlockContains
	_ = colorStatusStdoutV2
	_ = dirtyColorStatusRepo
}
```