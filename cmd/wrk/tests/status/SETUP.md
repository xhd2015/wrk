# Scenario

**Feature**: wrk --status displays discovered git repository status blocks

```
# cwd resolves to an effective git toplevel; status mode scans that root
wrk --status from cwd -> scan_repo.Scan(root) -> status blocks

# status is standalone; combining with another mode is rejected
wrk --status + other mode -> error (mutually exclusive)
```

## Preconditions

- Git must be available.
- `wrk --status` is a standalone mode.

## Steps

- Tests invoke `wrk --status` by default with `req.Args = []string{"--status"}`.
- Descendant scenarios choose whether cwd is inside a git checkout and whether another mode is also present.

## Context

- Successful status output is a sequence of blocks containing `Dir`, `Branch`, `Commit`, and `Status` lines.
- `Status` is `clean` or `dirty (N added, N changed, N renamed, N deleted)`; porcelain `??` untracked counts as **added** (same wrk taxonomy as `--projects`).
- The `Dir` line is relative to the current checkout toplevel; the checkout itself is `.`.
- **Main repo checkout cwd only**: the root `Dir: .` block also includes `Remote:` (same brief labels as `--projects`; `(no upstream)` when no tracking remote). Linked worktree cwd and nested `RepoTypeMain` repos omit `Remote:`.
- **Linked worktrees only** (`worktree.IsLinked`) also include one-line `Master:` — brief branch-relation label comparing the main repo's current branch vs the worktree's current branch (`git.CompareBranches`: `identical`, `needs merge back(+N commit(s))`, `needs fast forward(+N commit(s))`, `diverged(N commit(s))`); main checkout and nested independent `RepoTypeMain` repos omit this field.
- When stdout is a TTY or `--color` is set, `--status` colors `Status: clean` green and applies granular dirty-status coloring (same rules as `--projects`); `Master:` values use green/orange/red by relation. Without color: plain text.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	return nil
}

func statusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func statusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	branch := gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	return "Branch:       " + branch
}

func statusNoUpstreamRemote() string {
	return "Remote:       (no upstream)"
}

func statusBlockPlain(t *testing.T, repoDir, relDir, statusLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s",
		relDir, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir), statusLine)
}

func statusRootBlockPlain(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       %s\n%s",
		statusBranchLine(t, mainRepo), statusCommitLine(t, mainRepo), statusLine, remoteLine)
}

func statusBlockTemplate(t *testing.T, repoDir, relDir, statusLine string) string {
	t.Helper()
	if relDir == "." {
		return v2StdoutTemplate(statusRootBlockPlain(t, repoDir, statusLine, statusNoUpstreamRemote()))
	}
	return v2StdoutTemplate(statusBlockPlain(t, repoDir, relDir, statusLine))
}

func statusRootBlockTemplate(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return v2StdoutTemplate(statusRootBlockPlain(t, mainRepo, statusLine, remoteLine))
}

func statusStdoutV2(t *testing.T, blocks ...string) string {
	t.Helper()
	return v2StdoutTemplate(joinStdoutBlocks(blocks...))
}

func statusInitRepoWithSubject(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func statusOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}
```
