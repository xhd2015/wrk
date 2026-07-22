# Scenario

**Feature**: wrk --sync CLI (Phase 1 flags / Phase 2 WIP helper / Phase 3 full FF sync)

```
# isolated WRK_HOME + git or neutral cwd per test; build wrk once per session
git main-only repo -> wrk --sync [--dry-run] -> summary line (zeros) OR flag error
non-git / invalid flags -> non-zero stderr contains
IsWipSubject(subject) -> bool (Phase 2 probe mode)
main + linked worktrees -> wrk --sync [--dry-run]
  -> pass1 harvest (main ← wt) + pass2 distribute (wt ← main)
  -> detail lines when actions > 0; summary counts; warning: skips
```

## Preconditions

- The wrk Go module root is three levels above this test tree (`cmd/wrk/tests/sync/`).
- Go toolchain and git are available on PATH.
- Process-local wrk binary built once under an in-memory mutex (not session flock).
- Git helpers use `github.com/xhd2015/gitops/git/git_isolated` (hook-free).

## Context

- Each leaf gets isolated `{WorkRoot}` and `{WorkRoot}/.wrk` as `WRK_HOME`.
- `initMainOnlyRepo` seeds a single main checkout on branch `main` with one commit
  and **no** linked worktrees (Phase 1 success path).
- Phase 3 helpers: `initMainWithLinkedBranch`, `addLinkedWorktree`, `commitOnRepo`,
  `makeDirty`, `revParseHEAD`, `shortHEAD` for multi-worktree FF / skip fixtures.
- `runWrkSync` is satisfied by root `Run` + `req.Args`.
- Stdout assertions use `assert.Output` v2 full-match templates for bounded success output.
- Stderr errors/warnings use contains-match (`assertContains` or `assert.Output` `<contains>`)
  when hashes or absolute paths appear; exact when fully pinned.
- **Zero-summary contract (Phase 1):** when into-main=0, into-worktrees=0 and no
  detail actions, stdout is exactly one summary line (no blank/detail lines):
  - `synced: 0 into main, 0 into worktrees, 0 skipped\n`
  - `would: synced: 0 into main, 0 into worktrees, 0 skipped\n`
- **Detail lines only when actions > 0**, then a blank line, then the summary.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/doctest/session"
	"sync"
)

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

// Process-local wrk binary (one-process suite; in-memory mutex, not session.Once/flock).
var (
	wrkBinMu   sync.Mutex
	wrkBinPath string
	wrkBinErr  error
	// wrkModRoot set once via noteModRoot (sync.Once); not per-leaf Setup writes.
	wrkModOnce sync.Once
	wrkModRoot string
)


// noteModRoot records module root once per process (sync.Once). Safe under t.Parallel.
// Prefer this over writing wrkModRoot from every leaf Setup.
func noteModRoot(d *session.Doctest) {
	if d == nil {
		return
	}
	wrkModOnce.Do(func() {
		wrkModRoot = findModuleRoot(d.DOCTEST_ROOT)
	})
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	wrkBinMu.Lock()
	defer wrkBinMu.Unlock()
	if wrkBinPath != "" || wrkBinErr != nil {
		if wrkBinErr != nil {
			t.Fatal(wrkBinErr)
		}
		return wrkBinPath
	}
	if wrkModRoot == "" {
		t.Fatal("wrkModRoot unset; root Setup must call noteModRoot first")
	}
	dir, err := os.MkdirTemp("", "wrk-doctest-bin-")
	if err != nil {
		wrkBinErr = err
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "wrk")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/wrk")
	cmd.Dir = wrkModRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		wrkBinErr = fmt.Errorf("build wrk: %v\n%s", err, out)
		t.Fatal(wrkBinErr)
	}
	wrkBinPath = bin
	return bin
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	noteModRoot(d)
	if wrkModRoot == "" {
		t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
	}
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	syncEnsureHelpersUsed()
	return nil
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func commitFile(t *testing.T, repo, relPath, content, subject string) string {
	t.Helper()
	writeFile(t, filepath.Join(repo, relPath), content)
	runGitIsolated(t, repo, "add", relPath)
	runGitIsolated(t, repo, "commit", "-m", subject)
	return gitOutputIsolated(t, repo, "rev-parse", "HEAD")
}

func resolveRepoPath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

// initMainOnlyRepo: single main checkout on branch main, one commit, no linked worktrees.
func initMainOnlyRepo(t *testing.T, req *Request) string {
	t.Helper()
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# main only\n", "init")
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	return repo
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	mkdirAll(t, cwd)
	return cwd
}

// --- Phase 3 multi-worktree helpers ---

func revParseHEAD(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, dir, "rev-parse", "HEAD"))
}

func shortHEAD(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, dir, "rev-parse", "--short=7", "HEAD"))
}

func shortSHA(t *testing.T, dir, rev string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, dir, "rev-parse", "--short=7", rev))
}

func addLinkedWorktree(t *testing.T, mainRepo, branch, wtPath string) string {
	t.Helper()
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtPath)
	return resolveRepoPath(t, wtPath)
}

// initMainWithLinkedBranch: main on `main` (init commit) + one linked wt on `branch` at same tip.
func initMainWithLinkedBranch(t *testing.T, req *Request, branch, wtDirName string) (mainRepo, wtPath string) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	commitFile(t, mainRepo, "README.md", "# sync fixture\n", "init")
	mainRepo = resolveRepoPath(t, mainRepo)
	wtPath = addLinkedWorktree(t, mainRepo, branch, filepath.Join(req.WorkRoot, wtDirName))
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = branch
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = revParseHEAD(t, wtPath)
	return mainRepo, wtPath
}

func makeDirty(t *testing.T, repo, relPath, content string) {
	t.Helper()
	writeFile(t, filepath.Join(repo, relPath), content)
}

func assertHEAD(t *testing.T, dir, wantSHA string) {
	t.Helper()
	got := revParseHEAD(t, dir)
	if got != wantSHA {
		t.Fatalf("HEAD at %s: got %s want %s", dir, got, wantSHA)
	}
}

func assertHEADUnchanged(t *testing.T, dir, beforeSHA string) {
	t.Helper()
	assertHEAD(t, dir, beforeSHA)
}

// Phase 1 stable summary contracts (extend counts in later phases; keep shape).
const (
	syncSummaryZero      = "synced: 0 into main, 0 into worktrees, 0 skipped\n"
	syncWouldSummaryZero = "would: synced: 0 into main, 0 into worktrees, 0 skipped\n"
)

func commitWord(n int) string {
	if n == 1 {
		return "commit"
	}
	return "commits"
}

func syncSummaryLine(intoMain, intoWT, skipped int) string {
	return fmt.Sprintf("synced: %d into main, %d into worktrees, %d skipped\n", intoMain, intoWT, skipped)
}

func syncWouldSummaryLine(intoMain, intoWT, skipped int) string {
	return "would: " + syncSummaryLine(intoMain, intoWT, skipped)
}

// Detail line for pass-1 harvest: main ← <branch>  (+N commit(s))
func syncDetailPass1(branch string, n int) string {
	return fmt.Sprintf("main ← %s  (+%d %s)\n", branch, n, commitWord(n))
}

// Detail line for pass-2 distribute: <branch> ← main  (+N commit(s))
func syncDetailPass2(branch string, n int) string {
	return fmt.Sprintf("%s ← main  (+%d %s)\n", branch, n, commitWord(n))
}

// buildSyncStdout builds stdout with optional detail lines (actions > 0 only).
// When dryRun, every line is prefixed with "would: " (including the summary).
// When details is empty, body is exactly the summary line (Phase 1 zero-compat).
func buildSyncStdout(details []string, intoMain, intoWT, skipped int, dryRun bool) string {
	var b strings.Builder
	prefix := ""
	if dryRun {
		prefix = "would: "
	}
	for _, d := range details {
		line := strings.TrimSuffix(d, "\n") + "\n"
		b.WriteString(prefix)
		b.WriteString(line)
	}
	if len(details) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(prefix)
	b.WriteString(syncSummaryLine(intoMain, intoWT, skipped))
	return b.String()
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

func syncStdoutV2(body string) string {
	return v2StdoutTemplate(body)
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func assertEmptyStderr(t *testing.T, stderr string) {
	t.Helper()
	if stderr != "" {
		t.Fatalf("stderr should be empty, got %q", stderr)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func buildSyncCLIArgs(req *Request) []string {
	return append([]string(nil), req.Args...)
}

func syncWrkEnv(req *Request) []string {
	return append(os.Environ(), "WRK_HOME="+req.WrkHome)
}

func syncEnsureHelpersUsed() {
	_ = buildSyncCLIArgs
	_ = syncWrkEnv
	_ = initMainOnlyRepo
	_ = initNeutralCwd
	_ = initMainWithLinkedBranch
	_ = addLinkedWorktree
	_ = revParseHEAD
	_ = shortHEAD
	_ = shortSHA
	_ = makeDirty
	_ = assertHEAD
	_ = assertHEADUnchanged
	_ = commitWord
	_ = syncSummaryLine
	_ = syncWouldSummaryLine
	_ = syncDetailPass1
	_ = syncDetailPass2
	_ = buildSyncStdout
	_ = syncStdoutV2
	_ = assertEmptyStdout
	_ = assertEmptyStderr
	_ = assertContains
	_ = assertOutputExact
}
```
