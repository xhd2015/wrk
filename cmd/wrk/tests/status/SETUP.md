# Scenario

**Feature**: wrk --status displays discovered git repository status blocks

```
# cwd resolves to an effective git toplevel; status mode scans that root
wrk --status from cwd -> scan_repo.Scan(root) -> status blocks

# main-repo status: primary (main + ListLinked) then optional external section
main-repo --status
  -> PartitionStatusPaths(main, scan, ListLinked)
  -> primary blocks first
  -> if external non-empty: blank + "---- external ----" (+ gray when color) + blank + external blocks

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
- Every `Dir:` value uses **invocation cwd** (process work directory when wrk started):
  `filepath.Rel(normalize(cwd), normalize(repoPath))`; on Rel failure, or when the
  cleaned relative path has **more than two** leading `..` segments, print the
  absolute normalized path; otherwise print `filepath.ToSlash(rel)`.
  Examples: cwd at checkout root → `.`; cwd at `main/pkg/api` → main shows `../..`;
  cwd at `main/a/b/c/d` → main shows absolute (four leading `..`).
- **Main repo identity** (not `Dir == "."`): when statusing a main-repo checkout, the
  main-repo block includes `Remote:` (same brief labels as `--projects`; `(no upstream)`
  when no tracking remote). Linked worktree cwd and nested `RepoTypeMain` repos omit
  `Remote:`.
- **Main-repo sections (P2+P3)**: primary paths are main then `worktree.ListLinked` porcelain
  order (in-tree + out-of-tree + prunable). External paths are scan hits not in primary,
  path-sorted. When external is non-empty, after the last primary block print a blank
  line, the header line `---- external ----` (P3: gray ANSI `#90` when `colorEnabled`;
  plain ASCII when color off), another blank line, then external blocks. Omit the header
  entirely when external is empty.
- **Linked worktrees only** (`worktree.IsLinked`) also include one-line `Master:` — brief branch-relation label comparing the main repo's current branch vs the worktree's current branch (`git.CompareBranches`: `identical`, `needs merge back(+N commit(s))`, `needs fast forward(+N commit(s))`, `diverged(N commit(s))`); main checkout and nested independent `RepoTypeMain` repos omit this field.
- When stdout is a TTY or `--color` is set, `--status` colors `Status: clean` green and applies granular dirty-status coloring (same rules as `--projects`); `Master:` values use green/orange/red by relation; external section header is gray meta. Without color: plain text.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
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

// statusNormalizePath matches storage.NormalizePath / resolvePath used by product Dir abs fallback.
func statusNormalizePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// statusDirLine mirrors product statusDirLine(displayCwd, repoPath):
// rel = Rel(norm(cwd), norm(repo)); Rel fail or leading ".." count > 2 → absolute; else ToSlash(rel).
func statusDirLine(t *testing.T, invCwd, repoPath string) string {
	t.Helper()
	base := statusNormalizePath(t, invCwd)
	target := statusNormalizePath(t, repoPath)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	rel = filepath.Clean(rel)
	slash := filepath.ToSlash(rel)
	leading := 0
	for _, p := range strings.Split(slash, "/") {
		if p == ".." {
			leading++
			continue
		}
		break
	}
	if leading > 2 {
		return target
	}
	return slash
}

func statusDirField(t *testing.T, invCwd, repoPath string) string {
	t.Helper()
	return "Dir:          " + statusDirLine(t, invCwd, repoPath)
}

func statusBlockPlain(t *testing.T, repoDir, dirLine, statusLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s",
		dirLine, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir), statusLine)
}

// statusRootBlockWithDir builds a main-repo block. dirLine comes from statusDirLine (may be
// ".", "../..", or absolute). Remote is attached for main identity — not gated on Dir==".".
func statusRootBlockWithDir(t *testing.T, mainRepo, dirLine, statusLine, remoteLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s\n%s",
		dirLine, statusBranchLine(t, mainRepo), statusCommitLine(t, mainRepo), statusLine, remoteLine)
}

// statusRootBlockPlain is the main-root convenience form (Dir: .).
func statusRootBlockPlain(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return statusRootBlockWithDir(t, mainRepo, ".", statusLine, remoteLine)
}

// statusBlockTemplate: dirLine "." implies main-root block with Remote (historical helper).
// For non-dot main Dir lines (subdir cwd), use statusMainBlockFromCwd / statusRootBlockWithDir.
func statusBlockTemplate(t *testing.T, repoDir, dirLine, statusLine string) string {
	t.Helper()
	if dirLine == "." {
		return v2StdoutTemplate(statusRootBlockPlain(t, repoDir, statusLine, statusNoUpstreamRemote()))
	}
	return v2StdoutTemplate(statusBlockPlain(t, repoDir, dirLine, statusLine))
}

func statusRootBlockTemplate(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return v2StdoutTemplate(statusRootBlockPlain(t, mainRepo, statusLine, remoteLine))
}

// statusMainBlockFromCwd builds expected main-repo block (with Remote) using invocation-cwd Dir rule.
func statusMainBlockFromCwd(t *testing.T, invCwd, mainRepo, statusLine string) string {
	t.Helper()
	return statusRootBlockWithDir(t, mainRepo, statusDirLine(t, invCwd, mainRepo), statusLine, statusNoUpstreamRemote())
}

func statusStdoutV2(t *testing.T, blocks ...string) string {
	t.Helper()
	return v2StdoutTemplate(joinStdoutBlocks(blocks...))
}

// statusExternalSectionHeader is the plain section marker between primary and
// external status blocks (no ANSI; used when color is off).
func statusExternalSectionHeader() string {
	return "---- external ----"
}

// statusExternalSectionHeaderColored is the P3 gray meta header when colorEnabled
// (maps to product ansiGrey / #90; assert token "gray").
func statusExternalSectionHeaderColored() string {
	return "<ansi-color gray>---- external ----</ansi-color>"
}

// statusStdoutPrimaryExternal joins primary blocks, optional plain header, then
// external blocks with the same blank-line rhythm as inter-block separators.
// When external is empty, the header is omitted entirely.
func statusStdoutPrimaryExternal(t *testing.T, primary []string, external []string) string {
	t.Helper()
	return statusStdoutPrimaryExternalHeader(t, primary, external, false)
}

// statusStdoutPrimaryExternalColored is like statusStdoutPrimaryExternal but wraps
// the external section header in gray ANSI (P3; --color / TTY color path).
func statusStdoutPrimaryExternalColored(t *testing.T, primary []string, external []string) string {
	t.Helper()
	return statusStdoutPrimaryExternalHeader(t, primary, external, true)
}

func statusStdoutPrimaryExternalHeader(t *testing.T, primary []string, external []string, colorHeader bool) string {
	t.Helper()
	parts := make([]string, 0, len(primary)+1+len(external))
	parts = append(parts, primary...)
	if len(external) > 0 {
		if colorHeader {
			parts = append(parts, statusExternalSectionHeaderColored())
		} else {
			parts = append(parts, statusExternalSectionHeader())
		}
		parts = append(parts, external...)
	}
	return statusStdoutV2(t, parts...)
}

func assertNoExternalSectionHeader(t *testing.T, stdout string) {
	t.Helper()
	// Core text is present with or without gray ANSI wrappers.
	if strings.Contains(stdout, statusExternalSectionHeader()) {
		t.Fatalf("stdout must not contain %q, got:\n%s", statusExternalSectionHeader(), stdout)
	}
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
