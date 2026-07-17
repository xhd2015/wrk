# Scenario

**Feature**: wrk --status from main prints primary (main + ListLinked) then optional external section

```
# primary = main + ListLinked porcelain; WRK out-of-tree linked are primary (not the section header)
wrk --status from main cwd
  -> PartitionStatusPaths(main, scan, ListLinked)
  -> primary blocks (Dir via statusDirLine; out-of-tree/prunable keep appended-style presentation)

# section header only when scan has non-primary (nested) paths
# WRK out-of-tree linked are primary → no header for them alone
external non-empty -> blank + "---- external ----" (+ gray when color) + blank + external blocks
external empty -> no header

# linked worktree cwd (no --main) skips main-repo sectioning
wrk --status from external wt cwd -> scan only, no primary/external sections
```

## Preconditions

- Git must be available.
- Each test uses isolated `WRK_HOME` at `{WorkRoot}/.wrk` and `WRK_DATE=2026-06-30`.
- External worktrees are created via `wrk` (no args) from the main repo checkout.
- In-tree linked worktrees use `git worktree add` under the main repo tree.

## Steps

- Descendants set up repos/worktrees and set `req.RepoDir` (cwd for `--status`).
- Default `req.Args` is `[]string{"--status"}`; color tests add `--color`.

## Context

- **All** `Dir:` lines (primary + external + broken/prunable) use `statusDirLine(invCwd, repoPath)`:
  Rel from **invocation cwd**; Rel fail or leading `..` count > 2 → absolute normalized path;
  else `filepath.ToSlash(rel)`. Typical fixture external under `{WorkRoot}/.wrk/worktrees/…`
  from main root is often `../.wrk/worktrees/…` (one `..` → **relative**, not absolute).
- **Primary** includes main-owned linked worktrees (in-tree + WRK out-of-tree + prunable)
  in ListLinked porcelain order after main. These do **not** get a section header.
- Healthy primary linked blocks include `Branch`, `Commit`, `Status`, and `Master:`.
- Broken/prunable primary linked blocks are minimal two-line blocks (no Branch/Commit/Master).
- Multi-block stdout is joined with `\n\n` (blank line between every block and around the
  `---- external ----` header when external is non-empty; header is plain without
  color, gray ANSI when colorEnabled — P3).
- `Remote:` on the main block is gated by main identity, not by `Dir == "."`.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const wrkDate = "2026-06-30"

func findModuleRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func fixtureCacheBase(t *testing.T) string {
	t.Helper()
	base := os.Getenv("DOCTEST_FIXTURE_ROOT")
	if base != "" {
		return base
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(home, "Library", "Caches", "doctest", "fixtures")
}

func fixtureSessionRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixtureCacheBase(t), DOCTEST_SESSION_ID)
}

func sessionWrkBin(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixtureSessionRoot(t), "bin", "wrk")
}

func withFlock(t *testing.T, lockPath string, fn func()) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("mkdir lock dir: %v", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open lock %s: %v", lockPath, err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock %s: %v", lockPath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()
	fn()
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	bin := sessionWrkBin(t)
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	lockPath := filepath.Join(fixtureSessionRoot(t), "bin", ".lock")
	withFlock(t, lockPath, func() {
		if _, err := os.Stat(bin); err == nil {
			return
		}
		modRoot := findModuleRoot(DOCTEST_ROOT)
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors")
		}
		cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0755); err != nil {
		return err
	}
	if len(req.Args) == 0 {
		req.Args = []string{"--status"}
	}
	ensureMainRepoWtHelpersUsed()
	return nil
}

func wrkEnv(req *Request) []string {
	return append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func gitOutputIsolated(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git_isolated.MustOutput(t, dir, args...)
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func resolvePath(t *testing.T, path string) string {
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

// statusNormalizePath matches storage.NormalizePath / resolvePath used by product Dir abs fallback.
// Duplicated here: main-repo-worktrees is its own doctest root (own DOCTEST.md) and does not
// inherit helpers from cmd/wrk/tests/status/SETUP.md.
func statusNormalizePath(t *testing.T, path string) string {
	t.Helper()
	return resolvePath(t, path)
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

func v2StdoutTemplate(body string) string {
	if body == "" {
		return "---\nversion: 2\n---\n"
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return "---\nversion: 2\n---\n" + body
}

func joinStdoutBlocks(blocks ...string) string {
	trimmed := make([]string, 0, len(blocks))
	for _, b := range blocks {
		b = strings.TrimSuffix(b, "\n")
		if b != "" {
			trimmed = append(trimmed, b)
		}
	}
	result := strings.Join(trimmed, "\n\n")
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func statusStdoutV2(t *testing.T, blocks ...string) string {
	t.Helper()
	return v2StdoutTemplate(joinStdoutBlocks(blocks...))
}

// statusExternalSectionHeader is the plain (color-off) marker between primary and
// external sections. Product uses ansiGrey / #90 when colorEnabled (P3; covered in
// parent status color-output leaves, not this nested tree's no-color fixtures).
func statusExternalSectionHeader() string {
	return "---- external ----"
}

// statusStdoutPrimaryExternal joins primary blocks, optional plain header, then
// external blocks. Header is omitted when external is empty (e.g. only main +
// ListLinked / WRK out-of-tree primary linked).
func statusStdoutPrimaryExternal(t *testing.T, primary []string, external []string) string {
	t.Helper()
	parts := make([]string, 0, len(primary)+1+len(external))
	parts = append(parts, primary...)
	if len(external) > 0 {
		parts = append(parts, statusExternalSectionHeader())
		parts = append(parts, external...)
	}
	return statusStdoutV2(t, parts...)
}

func assertNoExternalSectionHeader(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, statusExternalSectionHeader()) {
		t.Fatalf("stdout must not contain %q, got:\n%s", statusExternalSectionHeader(), stdout)
	}
}

func statusOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}

// listLinkedPaths returns worktree.ListLinked paths in porcelain order.
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

// primaryLinkedBlockPlain builds the expected primary linked block for a path.
// Out-of-tree / dead paths use appended presentation; in-tree scan hits use scan block + Master.
func primaryLinkedBlockPlain(t *testing.T, invCwd, mainRepo, wtDir, wtBranch, statusLine string) string {
	t.Helper()
	mainNorm := resolvePath(t, mainRepo)
	wtNorm := resolvePath(t, wtDir)
	inTree := strings.HasPrefix(wtNorm, mainNorm+string(filepath.Separator))
	if worktree.IsDead(wtDir) {
		return appendedMinimalBlockPlain(t, invCwd, wtDir, "prunable")
	}
	if inTree {
		return scanStatusBlockFromCwd(t, invCwd, wtDir, statusLine, masterField(t, mainRepo, "main", wtBranch), false)
	}
	return appendedHealthyBlockPlain(t, invCwd, mainRepo, wtDir, wtBranch, statusLine)
}

func initMainRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func initMainRepoWithGoMod(t *testing.T, path, subject string) {
	t.Helper()
	initMainRepo(t, path, subject)
	writeFile(t, filepath.Join(path, "go.mod"), "module example.com/"+filepath.Base(path)+"\n\ngo 1.21\n")
	runGitIsolated(t, path, "add", "go.mod")
	runGitIsolated(t, path, "commit", "-m", "add go.mod")
}

func runWrkFrom(t *testing.T, req *Request, dir string) string {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin)
	cmd.Dir = dir
	cmd.Env = wrkEnv(req)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk exit %d stderr=%q", ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func branchName(baseBranch string, suffix int) string {
	name := baseBranch + "-" + wrkDate
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return name
}

func createExternalWrkWorktree(t *testing.T, req *Request) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	initMainRepoWithGoMod(t, mainRepo, "main repo for external wt")
	wtDir = runWrkFrom(t, req, mainRepo)
	branch = branchName("main", 0)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch
	return mainRepo, wtDir, branch
}

func createSecondExternalWrkWorktree(t *testing.T, req *Request, mainRepo string) (wtDir, branch string) {
	t.Helper()
	wtDir = runWrkFrom(t, req, mainRepo)
	branch = branchName("main", 1)
	req.WtDir2 = wtDir
	req.WtBranch2 = branch
	return wtDir, branch
}

func addInTreeLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func statusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func statusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

type porcelainCounts struct {
	added, changed, renamed, deleted int
}

func gitStatusCounts(t *testing.T, repoPath string) porcelainCounts {
	t.Helper()
	out := gitOutputIsolated(t, repoPath, "status", "--porcelain", "--untracked-files=no")
	var counts porcelainCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if len(line) < 2 {
			counts.changed++
			continue
		}
		x, y := line[0], line[1]
		switch {
		case x == 'R' || y == 'R':
			counts.renamed++
		case x == 'A' || y == 'A':
			counts.added++
		case x == 'D' || y == 'D':
			counts.deleted++
		default:
			counts.changed++
		}
	}
	return counts
}

func formatStatusLine(counts porcelainCounts) string {
	if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
		return "clean"
	}
	return fmt.Sprintf("dirty (%d added, %d changed, %d renamed, %d deleted)",
		counts.added, counts.changed, counts.renamed, counts.deleted)
}

func statusLineForRepo(t *testing.T, repoPath string) string {
	t.Helper()
	return formatStatusLine(gitStatusCounts(t, repoPath))
}

func statusNoUpstreamRemote() string {
	return "Remote:       (no upstream)"
}

// scanStatusBlockPlain builds one scan-phase block. dirLine is already computed via
// statusDirLine(invCwd, repoDir). withRemote is main-identity (not dirLine==".").
func scanStatusBlockPlain(t *testing.T, repoDir, dirLine, statusLine, masterLine string, withRemote bool) string {
	t.Helper()
	block := fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s",
		dirLine, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir), statusLine)
	if withRemote {
		block += "\n" + statusNoUpstreamRemote()
	}
	if masterLine != "" {
		block += "\n" + masterLine
	}
	return block
}

// scanStatusBlockFromCwd computes Dir from invCwd via statusDirLine.
func scanStatusBlockFromCwd(t *testing.T, invCwd, repoDir, statusLine, masterLine string, withRemote bool) string {
	t.Helper()
	return scanStatusBlockPlain(t, repoDir, statusDirLine(t, invCwd, repoDir), statusLine, masterLine, withRemote)
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

func appendedDirLine(t *testing.T, invCwd, wtPath string) string {
	t.Helper()
	return statusDirField(t, invCwd, wtPath)
}

func appendedHealthyBlockPlain(t *testing.T, invCwd, mainRepo, wtDir, wtBranch, statusLine string) string {
	t.Helper()
	master := masterField(t, mainRepo, "main", wtBranch)
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s",
		appendedDirLine(t, invCwd, wtDir),
		statusBranchLine(t, wtDir),
		statusCommitLine(t, wtDir),
		statusLine,
		master,
	)
}

func appendedMinimalBlockPlain(t *testing.T, invCwd, wtPath, statusLine string) string {
	t.Helper()
	return appendedDirLine(t, invCwd, wtPath) + "\nStatus:       " + statusLine
}

func appendedErrorStatusPlain(t *testing.T, wtPath string) string {
	t.Helper()
	return "error: " + worktreeGitError(t, wtPath)
}

func appendedErrorStatusColored(t *testing.T, wtPath string) string {
	t.Helper()
	gitErr := worktreeGitError(t, wtPath)
	return "<ansi-color red>error: " + gitErr + "</ansi-color>"
}

func gitCommandCombinedError(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git_isolated.Command(dir, args...).CombinedOutput()
	if err == nil {
		t.Fatalf("git %v in %s: expected failure", args, dir)
	}
	return strings.TrimSpace(string(out))
}

func worktreeGitError(t *testing.T, wtPath string) string {
	t.Helper()
	raw := gitCommandCombinedError(t, wtPath, "rev-parse", "--verify", "HEAD")
	return fmt.Sprintf("git rev-parse --verify HEAD in %s: %s", wtPath, raw)
}

func breakWorktreeGitMetadata(t *testing.T, req *Request, wtDir string) {
	t.Helper()
	staleGitdir := filepath.Join(req.WorkRoot, "stale-main", ".git", "worktrees", filepath.Base(wtDir))
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+staleGitdir+"\n")
}

func removeWorktreeCheckout(t *testing.T, wtPath string) {
	t.Helper()
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("remove worktree checkout %s: %v", wtPath, err)
	}
}

func dirtyWorktreeFile(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
}

func externalAppendOrder(t *testing.T, mainRepo string) []string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	scanRoots := map[string]bool{
		resolvePath(t, mainRepo): true,
	}
	var external []string
	for _, entry := range linked {
		norm := resolvePath(t, entry.Path)
		if scanRoots[norm] {
			continue
		}
		// In-tree linked worktrees live under mainRepo and are discovered by scan.
		if strings.HasPrefix(norm, resolvePath(t, mainRepo)+string(filepath.Separator)) {
			continue
		}
		external = append(external, entry.Path)
	}
	return external
}

func assertStdoutHasNoAppendedAbsDir(t *testing.T, stdout, absDir string) {
	t.Helper()
	if strings.Contains(stdout, "Dir:          "+absDir) {
		t.Fatalf("stdout should not contain appended absolute Dir %q, got:\n%s", absDir, stdout)
	}
}

func assertStdoutBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := statusOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d status blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between blocks, got:\n%s", stdout)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func ensureMainRepoWtHelpersUsed() {
	_ = getWrkBin
	_ = wrkEnv
	_ = resolvePath
	_ = statusNormalizePath
	_ = statusDirLine
	_ = statusDirField
	_ = statusStdoutV2
	_ = statusStdoutPrimaryExternal
	_ = statusExternalSectionHeader
	_ = assertNoExternalSectionHeader
	_ = listLinkedPaths
	_ = primaryLinkedBlockPlain
	_ = scanStatusBlockFromCwd
	_ = createExternalWrkWorktree
	_ = createSecondExternalWrkWorktree
	_ = addInTreeLinkedWorktree
	_ = appendedHealthyBlockPlain
	_ = appendedMinimalBlockPlain
	_ = breakWorktreeGitMetadata
	_ = externalAppendOrder
	_ = assertStdoutHasNoAppendedAbsDir
}
```