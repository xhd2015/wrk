# Scenario

**Feature**: wrk --projects detailed status output

```
wrk --projects -> one detailed status block per recorded main repo (lexicographic order)
```

## Steps

- Descendants vary whether any projects have been recorded.

## Context

- Each block uses absolute `Dir`, standard status fields, `Remote`, and `Worktrees:    N total, M dirty` (four spaces after colon).

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
	ensureProjectsHelpersUsed()
	return nil
}

func setupBareOriginForList(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func projectListBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func projectListCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func projectListCompareRemoteField(t *testing.T, mainRepo string) string {
	t.Helper()
	upstream := gitOutputIsolated(t, mainRepo, "rev-parse", "--abbrev-ref", "@{upstream}")
	if upstream == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstream, "main")
	if err != nil {
		t.Fatalf("CompareBranches: %v", err)
	}
	if result.Relation != git.BranchRelationSame {
		t.Fatalf("expected identical upstream, got relation %v", result.Relation)
	}
	return "Remote:       identical"
}

func projectListBlock(t *testing.T, mainRepo string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       clean\n%s\nWorktrees:    0 total, 0 dirty",
		resolvePath(t, mainRepo),
		projectListBranchLine(t, mainRepo),
		projectListCommitLine(t, mainRepo),
		projectListCompareRemoteField(t, mainRepo),
	)
}

func ensureProjectListHelpersUsed() {
	_ = setupBareOriginForList
	_ = projectListBranchLine
	_ = projectListCommitLine
	_ = projectListCompareRemoteField
	_ = projectListBlock
	_ = strings.TrimSpace
}
```