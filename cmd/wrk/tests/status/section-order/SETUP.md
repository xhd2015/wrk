# Scenario

**Feature**: main-repo `wrk --status` prints primary blocks first, then optional `---- external ----`

```
# P2 CLI wiring of PartitionStatusPaths + section header (plain without --color)
main-repo --status
  -> primary = main + ListLinked (porcelain)
  -> if external non-empty: blank + "---- external ----" + blank + external (path-sorted)
  -> if external empty: no header line
```

## Preconditions

- Git available; isolated `WRK_HOME` at `{WorkRoot}/.wrk` (parent harness).
- Section-order leaves run without `--color` (plain header). Gray header is covered under
  `color-output/force-color-header` (P3).

## Steps

- Leaves build main ± linked ± nested fixtures and set `req.RepoDir` / `req.Args`.
- Default args remain `["--status"]` from parent status SETUP.

## Context

- Header is plain ASCII when color is off (these leaves); gray ANSI when color on (P3, color-output).
- Main-owned WRK out-of-tree linked worktrees are **primary**, not the external section.
- Nested independent repos under the main tree are **external**.

```go
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	ensureSectionOrderHelpersUsed()
	return nil
}

func addInTreeLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func initNestedIndependentRepo(t *testing.T, mainRepo, relPath, subject string) string {
	t.Helper()
	// Ensure parent ignores nested path so root stays clean.
	child := filepath.Join(mainRepo, filepath.FromSlash(relPath))
	statusInitRepoWithSubject(t, child, subject)
	return child
}

func ensureToolsGitignore(t *testing.T, mainRepo string, patterns ...string) {
	t.Helper()
	content := strings.Join(patterns, "\n") + "\n"
	writeFile(t, filepath.Join(mainRepo, ".gitignore"), content)
	runGitIsolated(t, mainRepo, "add", ".gitignore")
	runGitIsolated(t, mainRepo, "commit", "-m", "ignore nested for clean parent porcelain")
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

func listLinkedPaths(t *testing.T, mainRepo string) []string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	paths := make([]string, 0, len(linked))
	for _, entry := range linked {
		paths = append(paths, entry.Path)
	}
	return paths
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	return statusNormalizePath(t, path)
}

func appendedHealthyBlockPlain(t *testing.T, invCwd, mainRepo, wtDir, wtBranch, statusLine string) string {
	t.Helper()
	master := masterField(t, mainRepo, "main", wtBranch)
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s",
		statusDirField(t, invCwd, wtDir),
		statusBranchLine(t, wtDir),
		statusCommitLine(t, wtDir),
		statusLine,
		master,
	)
}

func scanLinkedBlockFromCwd(t *testing.T, invCwd, mainRepo, wtDir, wtBranch, statusLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s\n%s",
		statusDirLine(t, invCwd, wtDir),
		statusBranchLine(t, wtDir),
		statusCommitLine(t, wtDir),
		statusLine,
		masterField(t, mainRepo, "main", wtBranch),
	)
}

func primaryLinkedBlockPlain(t *testing.T, invCwd, mainRepo, wtDir, wtBranch, statusLine string) string {
	t.Helper()
	mainNorm := resolvePath(t, mainRepo)
	wtNorm := resolvePath(t, wtDir)
	inTree := strings.HasPrefix(wtNorm, mainNorm+string(filepath.Separator))
	if inTree {
		return scanLinkedBlockFromCwd(t, invCwd, mainRepo, wtDir, wtBranch, statusLine)
	}
	return appendedHealthyBlockPlain(t, invCwd, mainRepo, wtDir, wtBranch, statusLine)
}

func createSecondExternalWrkWorktree(t *testing.T, req *Request, mainRepo string) (wtDir, branch string) {
	t.Helper()
	wtDir = runWrkFrom(t, req, mainRepo)
	branch = branchName("main", wrkDate, 1)
	req.Wt2Dir = wtDir
	req.Wt2Branch = branch
	return wtDir, branch
}

func ensureSectionOrderHelpersUsed() {
	_ = addInTreeLinkedWorktree
	_ = initNestedIndependentRepo
	_ = ensureToolsGitignore
	_ = masterField
	_ = listLinkedPaths
	_ = appendedHealthyBlockPlain
	_ = primaryLinkedBlockPlain
	_ = createSecondExternalWrkWorktree
}
```
