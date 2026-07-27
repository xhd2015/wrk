# Scenario

**Feature**: wrk --status adds brief Master: field on linked worktrees only

```
# scan discovers main checkout and in-tree linked worktrees
wrk --status from main cwd -> scan_repo.Scan(root) -> status blocks

# linked worktree blocks compare main-repo branch vs worktree branch (one-line brief label)
linked wt block -> Master: <identical|needs merge back|needs fast forward|diverged>

# main checkout and nested independent repos omit the field
main / nested RepoTypeMain blocks -> no Master: line
```

## Preconditions

- Git must be available.
- Linked worktrees are created inside the checkout root so `scan_repo.Scan` discovers them.
- `Master:` uses shared branch-relation brief labels (lowercase).

## Steps

- Descendants set up main repo + optional linked worktrees or nested repos, then run `wrk --status`.

## Context

- `Master:` compares the **main repo's current branch** (refA) against the **linked worktree's current branch** (refB).
- Only blocks where `worktree.IsLinked(repoPath)` is true include the field.
- One-line format: `Master:       <brief label>` (14-char aligned field prefix).

```go
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	return nil
}

func setupMainRepoWithSubject(t *testing.T, workRoot, name, subject string) string {
	t.Helper()
	path := filepath.Join(workRoot, name)
	statusInitRepoWithSubject(t, path, subject)
	return path
}

func addLinkedWorktreeInRepo(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func commitOnMain(t *testing.T, mainRepo, filename, content, subject string) {
	t.Helper()
	writeFile(t, filepath.Join(mainRepo, filename), content)
	runGitIsolated(t, mainRepo, "add", filename)
	runGitIsolated(t, mainRepo, "commit", "-m", subject)
}

func commitOnWorktree(t *testing.T, wtDir, filename, content, subject string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
	runGitIsolated(t, wtDir, "add", filename)
	runGitIsolated(t, wtDir, "commit", "-m", subject)
}

func masterBriefFromResult(result *git.CompareBranchesResult) string {
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

func masterField(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	return "Master:       " + masterBriefFromResult(result)
}

func statusBlockWithMasterPlain(t *testing.T, repoDir, relDir, statusLine, masterLine string) string {
	t.Helper()
	var block string
	if relDir == "." {
		block = statusRootBlockPlain(t, repoDir, statusLine, statusNoUpstreamRemote())
	} else {
		block = statusBlockPlain(t, repoDir, relDir, statusLine)
	}
	if masterLine != "" {
		block += "\n" + masterLine
	}
	return block
}

func statusBlockWithMaster(t *testing.T, repoDir, relDir, statusLine, masterLine string) string {
	t.Helper()
	return v2StdoutTemplate(statusBlockWithMasterPlain(t, repoDir, relDir, statusLine, masterLine))
}

func assertNoMasterField(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "Master:") {
		t.Fatalf("stdout should not contain Master:, got:\n%s", stdout)
	}
}

func ensureMasterFieldHelpersUsed() {
	_ = setupMainRepoWithSubject
	_ = addLinkedWorktreeInRepo
	_ = commitOnMain
	_ = commitOnWorktree
	_ = masterBriefFromResult
	_ = masterField
	_ = statusBlockWithMasterPlain
	_ = statusBlockWithMaster
	_ = assertNoMasterField
}
```