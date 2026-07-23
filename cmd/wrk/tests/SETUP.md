# Scenario

**Feature**: wrk CLI auto worktree, merge-back, and worktree listing

```
# isolated WRK_HOME + work root per test; build wrk once
wrk (no args) from cwd -> stdout path only + git worktree side effects
wrk --done [--confirm [--confirm-from-stdin]] from linked wt -> merge-back --rm
wrk --list from cwd -> git worktree list stdout unchanged
```

## Preconditions

- The wrk Go module root is three levels above the test tree root (at `cmd/wrk/tests/`)
- Go toolchain is available on PATH
- Git is required for worktree tests

## Context

Each test runs the `wrk` CLI in an isolated environment. The `wrk` binary is built once per process under an in-memory mutex into `os.MkdirTemp` (one-process suite; not session flock). Each leaf gets its own temp directory and isolated `WRK_HOME` at `{WorkRoot}/.wrk`.

```go
import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/wrk/wrkcli"
)

const (
	fixtureSeedMainReadme = "main-readme"
	fixtureSeedMainGoMod  = "main-gomod"
)

type seedBuilder func(seedDir string)

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

// Process-local seed repos (one-process; in-memory mutex, not session flock).
var (
	seedMu   sync.Mutex
	seedDirs = map[string]string{}
)

func ensureSeed(t *testing.T, seedID string, build seedBuilder) string {
	t.Helper()
	seedMu.Lock()
	defer seedMu.Unlock()
	if seedDir, ok := seedDirs[seedID]; ok && isValidGitRepo(seedDir) {
		if resolved, err := filepath.EvalSymlinks(seedDir); err == nil {
			seedDir = resolved
		}
		return seedDir
	}
	base, err := os.MkdirTemp("", "wrk-doctest-seeds-")
	if err != nil {
		t.Fatalf("mkdir seeds base: %v", err)
	}
	seedDir := filepath.Join(base, seedID)
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatalf("mkdir seed %s: %v", seedDir, err)
	}
	build(seedDir)
	if resolved, err := filepath.EvalSymlinks(seedDir); err == nil {
		seedDir = resolved
	}
	if !isValidGitRepo(seedDir) {
		t.Fatalf("seed %q not built", seedID)
	}
	seedDirs[seedID] = seedDir
	return seedDir
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	noteModRoot(d)
	if wrkModRoot == "" {
		t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
	}
	// Resolve symlinks so derived paths match git's resolved output (macOS
	// serves /var from /private/var; t.TempDir returns the symlinked form).
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	ensureHelpersUsed()
	return os.MkdirAll(req.WrkHome, 0755)
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

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
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

func gitWorktreeListIsolated(t *testing.T, dir string) string {
	t.Helper()
	return git_isolated.WorktreeList(t, dir)
}

func isValidGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func writeFileSeed(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(fmt.Sprintf("mkdir %s: %v", filepath.Dir(path), err))
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(fmt.Sprintf("write %s: %v", path, err))
	}
}

func runGitSeed(dir string, args ...string) {
	if err := git_isolated.Run(dir, args...); err != nil {
		panic(err.Error())
	}
}

func buildSeedMainReadme(seedDir string) {
	if err := git_isolated.Init(seedDir, "main"); err != nil {
		panic(err.Error())
	}
	runGitSeed(seedDir, "config", "user.email", git_isolated.DefaultUserEmail)
	runGitSeed(seedDir, "config", "user.name", git_isolated.DefaultUserName)
	writeFileSeed(filepath.Join(seedDir, "README.md"), "# test\n")
	runGitSeed(seedDir, "add", "README.md")
	runGitSeed(seedDir, "commit", "-m", "init")
}

func cloneMainGoModFromSeed(t *testing.T, dst string) {
	t.Helper()
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, dst)
}

func buildSeedMainGoMod(seedDir string) {
	buildSeedMainReadme(seedDir)
	writeFileSeed(filepath.Join(seedDir, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitSeed(seedDir, "add", "go.mod")
	runGitSeed(seedDir, "commit", "-m", "add go.mod")
}

func cloneDirCoW(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("dest already exists: %s", dst)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("cp", "-cR", src, dst).Run()
	}
	return exec.Command("cp", "-a", src, dst).Run()
}

func cloneRepoFromSeed(t *testing.T, seedID string, build seedBuilder, dest string) {
	t.Helper()
	seed := ensureSeed(t, seedID, build)
	if err := cloneDirCoW(seed, dest); err != nil {
		t.Fatalf("clone seed %q -> %q: %v", seedID, dest, err)
	}
	resolved, err := filepath.EvalSymlinks(dest)
	if err == nil {
		dest = resolved
	}
	if !isValidGitRepo(dest) {
		t.Fatalf("cloned repo %q is not a valid git repo", dest)
	}
}

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	cloneRepoFromSeed(t, fixtureSeedMainReadme, buildSeedMainReadme, path)
}

const wrkDate = "2026-06-30"

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func worktreePath(wrkHome, basename, token, date string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", basename, token, date)
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(wrkHome, "worktrees", name)
}

func branchName(baseBranch, date string, suffix int) string {
	name := baseBranch + "-" + date
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return name
}

func runWrkFrom(t *testing.T, req *Request, dir string) string {
	t.Helper()
	// Create entry is wrk --new (bare no-args is dashboard).
	return runWrkWithArgs(t, req, dir, "--new")
}

func runWrkWithArgs(t *testing.T, req *Request, dir string, args ...string) string {
	t.Helper()
	bin := getWrkBin(t)
	// Empty args would enter dashboard; fixture helpers that mean "create" must pass --new.
	if len(args) == 0 {
		args = []string{"--new"}
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk %v exit %d stderr=%q", args, ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertGitFileIsWorktreeLink(t *testing.T, wtDir string) {
	t.Helper()
	gitPath := filepath.Join(wtDir, ".git")
	info, err := os.Stat(gitPath)
	if os.IsNotExist(err) {
		t.Fatalf("%s should exist", gitPath)
	}
	if err != nil {
		t.Fatalf("stat %s: %v", gitPath, err)
	}
	if info.IsDir() {
		t.Fatalf("expected %s to be a regular file (linked worktree), got directory", gitPath)
	}
}

func assertBranchExists(t *testing.T, repoDir, branch string) {
	t.Helper()
	if err := git_isolated.Command(repoDir, "rev-parse", "--verify", "refs/heads/"+branch).Run(); err != nil {
		t.Fatalf("branch %q should exist in %s", branch, repoDir)
	}
}

func assertWorktreeListContains(t *testing.T, repoDir, wantPath string) {
	t.Helper()
	list := gitOutputIsolated(t, repoDir, "worktree", "list", "--porcelain")
	found := false
	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			p := strings.TrimPrefix(line, "worktree ")
			if p == wantPath {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("git worktree list should contain %q, got:\n%s", wantPath, list)
	}
}

func assertBranchCheckedOutInWorktree(t *testing.T, wtDir, wantBranch string) {
	t.Helper()
	got := gitOutputIsolated(t, wtDir, "rev-parse", "--abbrev-ref", "HEAD")
	if got != wantBranch {
		t.Fatalf("worktree %s: expected branch %q, got %q", wtDir, wantBranch, got)
	}
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

func assertStdoutExactPath(t *testing.T, stdout, wantPath string) {
	t.Helper()
	assert.Output(t, stdout, v2StdoutTemplate(wantPath))
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in %q", substr, s)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	}
}

func assertBranchNotExists(t *testing.T, repoDir, branch string) {
	t.Helper()
	if err := git_isolated.Command(repoDir, "rev-parse", "--verify", "refs/heads/"+branch).Run(); err == nil {
		t.Fatalf("branch %q should not exist in %s", branch, repoDir)
	}
}

func assertWorktreeListNotContains(t *testing.T, repoDir, wantPath string) {
	t.Helper()
	list := gitOutputIsolated(t, repoDir, "worktree", "list", "--porcelain")
	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			p := strings.TrimPrefix(line, "worktree ")
			if p == wantPath {
				t.Fatalf("git worktree list should not contain %q", wantPath)
			}
		}
	}
}

// setupWrkWorktreeFromMain creates myrepo on main, runs wrk once, and records paths on req.
func setupWrkWorktreeFromMain(t *testing.T, req *Request) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneMainGoModFromSeed(t, mainRepo)
	wtDir = runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	branch = branchName("main", wrkDate, 0)
	req.WtBranch = branch
	return mainRepo, wtDir, branch
}

func commitAheadOnWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
	runGitIsolated(t, wtDir, "add", filename)
	runGitIsolated(t, wtDir, "commit", "-m", "worktree commit")
}

// --- composition helpers: --done/--merge-back + --sync ---

func revParseHEAD(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, dir, "rev-parse", "HEAD"))
}

func assertHEAD(t *testing.T, dir, wantSHA string) {
	t.Helper()
	got := revParseHEAD(t, dir)
	if got != wantSHA {
		t.Fatalf("HEAD at %s: got %s want %s", dir, got, wantSHA)
	}
}

func assertEmptyStderr(t *testing.T, stderr string) {
	t.Helper()
	if stderr != "" {
		t.Fatalf("stderr should be empty, got %q", stderr)
	}
}

func syncCommitWord(n int) string {
	if n == 1 {
		return "commit"
	}
	return "commits"
}

// syncDetailPass2: pass-2 distribute line  <branch> ← main  (+N commit(s))\n
func syncDetailPass2(branch string, n int) string {
	return fmt.Sprintf("%s ← main  (+%d %s)\n", branch, n, syncCommitWord(n))
}

func syncSummaryLine(intoMain, intoWT, skipped int) string {
	return fmt.Sprintf("synced: %d into main, %d into worktrees, %d skipped\n", intoMain, intoWT, skipped)
}

// buildSyncStdout builds standalone sync stdout (details only when actions > 0).
func buildSyncStdout(details []string, intoMain, intoWT, skipped int) string {
	var b strings.Builder
	for _, d := range details {
		line := strings.TrimSuffix(d, "\n") + "\n"
		b.WriteString(line)
	}
	if len(details) > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(syncSummaryLine(intoMain, intoWT, skipped))
	return b.String()
}

// primaryThenSyncStdout joins primary MergeBack message and sync block with a blank line.
func primaryThenSyncStdout(primaryMsg string, details []string, intoMain, intoWT, skipped int) string {
	primary := strings.TrimSuffix(primaryMsg, "\n") + "\n"
	return primary + "\n" + buildSyncStdout(details, intoMain, intoWT, skipped)
}

func compositionResolvePath(t *testing.T, path string) string {
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

// setupCompositionTwoWTs creates:
//   - wrk-managed wtA (req.WtDir / req.WtBranch) with an ahead commit (feature-work)
//   - second linked wtB on branch feature-stays at the pre-ahead tip (req.Wt2Dir / req.Wt2Branch)
// RepoDir is set to wtA. Caller sets Args.
func setupCompositionTwoWTs(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtA, _ := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo

	wt2Path := filepath.Join(req.WorkRoot, "wt-stays")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "feature-stays", wt2Path)
	wt2Path = compositionResolvePath(t, wt2Path)
	req.Wt2Dir = wt2Path
	req.Wt2Branch = "feature-stays"

	commitAheadOnWorktree(t, wtA, "feature-work", "ahead of main")
	req.RepoDir = wtA
}

const cascadeAheadExternalDepModule = "example.com/dep"

// setupConsumerWithAheadExternalDep creates a consumer linked wt plus an external
// dep worktree with a commit ahead of dep main (cascade confirmation required).
func setupConsumerWithAheadExternalDep(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, mainRepo, "edit", "-require="+cascadeAheadExternalDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadeAheadExternalDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	req.ExternalWtDir = externalPath

	// Commit consumer porcelain after --dep (replace/gitignore) so D2 cascade
	// preflight dirty checks pass; replace stays in go.mod for replace-guard tests.
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "commit dep replace and ignore", "--allow-empty")

	writeFile(t, filepath.Join(externalPath, "dep.go"), "package dep // ahead fix\n")
	runGitIsolated(t, externalPath, "add", "dep.go")
	runGitIsolated(t, externalPath, "commit", "-m", "dep fix on external worktree")
}

// prepareAheadExternalDepConsumerForDone drops the consumer's local replace after
// setupConsumerWithAheadExternalDep and commits so consumer --done can finish
// after cascade auto-merges the ahead external dep (no extra-repo replace guard).
//
// wrk --dep leaves the consumer dirty (.gitignore /external, go.mod replace, often
// go.sum). --done requires a clean worktree when removing, so stage+commit all
// consumer changes (not go.mod alone).
func prepareAheadExternalDepConsumerForDone(t *testing.T, req *Request) {
	t.Helper()
	wtDir := req.WtDir
	runGoModInDir(t, wtDir, "edit", "-dropreplace="+cascadeAheadExternalDepModule)
	runGoModInDir(t, wtDir, "edit", "-droprequire="+cascadeAheadExternalDepModule)
	// Stage every consumer change left by --dep + dropreplace (go.mod, .gitignore,
	// go.sum if present, etc.). Do not leave porcelain dirty for --done.
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "drop dep replace for done")
	if porcelain := strings.TrimSpace(gitOutputIsolated(t, wtDir, "status", "--porcelain")); porcelain != "" {
		t.Fatalf("consumer worktree must be clean after prepare for --done; porcelain:\n%s", porcelain)
	}
}

func runGoModInDir(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}

func buildWrkCLIArgs(req *Request) []string {
	var args []string
	if req.TargetDir != "" {
		args = append(args, req.TargetDir)
	}
	if req.SpawnDir != "" {
		args = append(args, req.SpawnDir)
	}
	if req.TaskDesc != "" {
		taskFlag := req.TaskFlag
		if taskFlag == "" {
			taskFlag = "--task"
		}
		args = append(args, taskFlag, req.TaskDesc)
	}
	if req.SetTaskDesc != "" {
		args = append(args, "--set-task", req.SetTaskDesc)
	}
	args = append(args, req.Args...)
	return args
}

// needsProcessIsolation is true when the leaf requires child-process isolation
// (cmd.Env / cmd.Dir). Parallel-safe; matches DOCTEST_LINT.md §1.
// In-process RunCLI must not Setenv/Chdir/stdin-hijack for these.
func needsProcessIsolation(req *Request) bool {
	if req.StdinInput != "" {
		return true
	}
	if len(req.ExtraEnv) > 0 {
		return true
	}
	if req.PathPrepend != "" || req.FakeShellDir != "" {
		return true
	}
	if req.FakeHome != "" || req.UseMinimalPath {
		return true
	}
	if req.UseFollowupEnv || req.FollowupFile != "" {
		return true
	}
	if req.SetTaskEnv != "" || req.BasenameEnv != "" {
		return true
	}
	if req.ProjectsPerfLog != "" {
		return true
	}
	if req.ShellEnv != "" || req.FakeShellLog != "" || req.FakeShellExit != 0 {
		return true
	}
	return false
}

// needsInProcessCaptureOK is true when the leaf's args are modes whose product
// paths write via ctx.out()/errw (parallel-safe in-process capture).
// Compose primaries (done/merge-back/pipelines) and other unthreaded modes stay
// on the binary even if they also pass a capture-safe modifier (e.g. --sync).
func needsInProcessCaptureOK(req *Request) bool {
	args := buildWrkCLIArgs(req)
	// Empty args = dashboard (still process stdout until dashboard is threaded).
	if len(args) == 0 {
		return false
	}
	// --main rewrites workDir via ResolveMainRepo + more Getwd-adjacent edges;
	// keep on binary until that path is fully origWd-based.
	for _, a := range args {
		if a == "--main" {
			return false
		}
	}
	// Isolation-heavy or still-process-stream primaries → binary (L3).
	// Presence of any of these wins over capture-safe modifiers below.
	for _, a := range args {
		switch a {
		case "--done", "--merge-back", "--dep", "--bring", "--all-deps",
			"--cd", "--set-task",
			"--scan-git-repos", "--reinstall-local", "--bash-integration",
			"--gen-commit-msg", "--new", "--dashboard":
			return false
		}
	}
	// --skill --install prints via skills/install process stdout → binary.
	hasSkill, skillInstall := false, false
	for _, a := range args {
		if a == "--skill" {
			hasSkill = true
		}
		if a == "--install" {
			skillInstall = true
		}
	}
	if hasSkill && skillInstall {
		return false
	}
	// Pure modes known to print via ctx.out()/errw (and free modifiers like
	// --dry-run/--json/--color/--fetch/--push with tag-next compose).
	for _, a := range args {
		switch a {
		case "--status", "--repos", "--list":
			return true
		case "--sync", "--projects", "--where":
			return true
		case "--tag-next", "--propagate-tags", "--push":
			return true
		case "--add", "--rm", "--remove":
			return true
		case "--set-config", "--skill":
			return true
		case "--version", "-h", "--help":
			return true
		}
	}
	// Default: binary until mode is known capture-safe.
	return false
}

// runWrkBinary runs the product binary with cmd.Env/cmd.Dir only (L3-capable,
// Parallel-safe). Used for process-boundary leaves and any leaf that needs env
// isolation.
func runWrkBinary(t *testing.T, req *Request, args []string) (*Response, error) {
	t.Helper()
	bin := getWrkBin(t)
	if req.UseScriptTTY {
		return execScriptTTYWrk(t, req, bin, args)
	}
	if req.WebProbe {
		return runWebProbe(t, req, bin, args)
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)
	return captureCommandOutput(cmd, req.StdinInput)
}

// runWrkInProcess runs wrk via wrkcli.RunCLI (L2). Parallel-safe: only
// Stdout/Stderr/Dir/WrkHome/WrkDate — no ExtraEnv/PathPrepend/Stdin.
// Product output is captured only on paths that write via ctx.out()/ctx.errw().
func runWrkInProcess(t *testing.T, req *Request, args []string) (*Response, error) {
	t.Helper()
	if needsProcessIsolation(req) {
		t.Fatal("runWrkInProcess: leaf needs process isolation; use runWrkBinary (cmd.Env/cmd.Dir)")
	}
	var stdout, stderr bytes.Buffer
	opts := wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
		WrkDate: wrkDate,
	}
	code := wrkcli.RunCLI(args, opts)
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}, nil
}

func execScriptTTYWrk(t *testing.T, req *Request, bin string, args []string) (*Response, error) {
	t.Helper()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		scriptArgs := append([]string{"-q", "/dev/null", bin}, args...)
		cmd = exec.Command("script", scriptArgs...)
	case "linux":
		quoted := make([]string, len(args))
		for i, a := range args {
			quoted[i] = "'" + strings.ReplaceAll(a, "'", `'"'"'`) + "'"
		}
		shellCmd := bin + " " + strings.Join(quoted, " ")
		cmd = exec.Command("script", "-q", "-c", shellCmd, "/dev/null")
	default:
		t.Skipf("script TTY helper not implemented for %s", runtime.GOOS)
	}
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)
	return captureCommandOutput(cmd, req.StdinInput)
}

func captureCommandOutput(cmd *exec.Cmd, stdinInput string) (*Response, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if stdinInput != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(stdin, stdinInput); err != nil {
			return nil, err
		}
		stdin.Close()
		err = cmd.Wait()
		exitCode := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			} else {
				return nil, err
			}
		}
		return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, nil
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, nil
}

// slugify converts a task description into a path-safe slug.
// Rules: lowercase, non-letter-non-digit → "-", collapse runs of "-",
// trim leading/trailing "-", truncate to 64 runes.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	runes := []rune(s)
	if len(runes) > 64 {
		s = string(runes[:64])
	}
	s = strings.Trim(s, "-")
	return s
}

func worktreePathWithTask(wrkHome, basename, token, date, slug string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", basename, token, date)
	if slug != "" {
		name = fmt.Sprintf("%s-%s", name, slug)
	}
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(wrkHome, "worktrees", name)
}

func branchNameWithTask(baseBranch, date, slug string, suffix int) string {
	name := baseBranch + "-" + date
	if slug != "" {
		name = name + "-" + slug
	}
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return name
}

// createWorktreeWithTask is like setupWrkWorktreeFromMain but with a task description.
func wrkEnv(req *Request) []string {
	if req.UseMinimalPath {
		home := req.FakeHome
		if home == "" {
			home = req.WorkRoot
		}
		// Prefer WorkRoot/minimal-bin (no system git-lfs). Fall back to
		// /usr/bin:/bin only when the helper was not prepared.
		path := filepath.Join(req.WorkRoot, "minimal-bin")
		if st, err := os.Stat(path); err != nil || !st.IsDir() {
			path = "/usr/bin:/bin"
		}
		env := []string{
			"HOME=" + home,
			"PATH=" + path,
			"WRK_HOME=" + req.WrkHome,
			"WRK_DATE=" + wrkDate,
		}
		if req.SetTaskEnv != "" {
			env = append(env, req.SetTaskEnv)
		}
		if req.BasenameEnv != "" {
			env = append(env, req.BasenameEnv)
		}
		if req.ProjectsPerfLog != "" {
			env = append(env, "WRK_PROJECTS_PERF_LOG="+req.ProjectsPerfLog)
		}
		env = appendExtraEnv(env, req)
		env = appendCDEnv(env, req)
		return env
	}
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	if req.FakeHome != "" {
		env = append(env, "HOME="+req.FakeHome)
	}
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	if req.BasenameEnv != "" {
		env = append(env, req.BasenameEnv)
	}
	if req.ProjectsPerfLog != "" {
		env = append(env, "WRK_PROJECTS_PERF_LOG="+req.ProjectsPerfLog)
	}
	env = appendExtraEnv(env, req)
	env = appendCDEnv(env, req)
	return env
}

// appendExtraEnv adds ExtraEnv KEY=VAL entries (create-ux mocks, etc.).
func appendExtraEnv(env []string, req *Request) []string {
	if len(req.ExtraEnv) == 0 {
		return env
	}
	return append(env, req.ExtraEnv...)
}

// prependPATH inserts dir at the front of PATH in env (or appends PATH=dir:…).
func prependPATH(env []string, dir string) []string {
	if dir == "" {
		return env
	}
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + dir + string(os.PathListSeparator) + strings.TrimPrefix(e, "PATH=")
			return env
		}
	}
	return append(env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// appendCDEnv adds --cd test harness env: WRK_FOLLOWUP_FILE, fake-shell PATH/SHELL,
// PathPrepend (fake agent-run / tools), and WRK_FAKE_SHELL_* for the shim
// that LoginInteractive must resolve without hanging.
func appendCDEnv(env []string, req *Request) []string {
	if req.UseFollowupEnv && req.FollowupFile != "" {
		env = append(env, "WRK_FOLLOWUP_FILE="+req.FollowupFile)
	}
	// PathPrepend first so UX fakes win over FakeShellDir when both set.
	env = prependPATH(env, req.PathPrepend)
	env = prependPATH(env, req.FakeShellDir)
	if req.ShellEnv != "" {
		env = append(env, "SHELL="+req.ShellEnv)
	}
	if req.FakeShellLog != "" {
		env = append(env, "WRK_FAKE_SHELL_LOG="+req.FakeShellLog)
	}
	if req.FakeShellExit != 0 {
		env = append(env, fmt.Sprintf("WRK_FAKE_SHELL_EXIT=%d", req.FakeShellExit))
	}
	return env
}

func createWorktreeWithTask(t *testing.T, req *Request, taskDesc string) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneMainGoModFromSeed(t, mainRepo)
	slug := slugify(taskDesc)
	wtDir = runWrkWithArgs(t, req, mainRepo, "--task", taskDesc)
	req.WtDir = wtDir
	branch = branchNameWithTask("main", wrkDate, slug, 0)
	req.WtBranch = branch
	req.TaskDesc = taskDesc
	return mainRepo, wtDir, branch
}

// freeLocalPort returns an unused TCP port on 127.0.0.1 for --web --port tests.
func freeLocalPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeLocalPort: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

// lockingWriter guards concurrent writes to a bytes.Buffer (stdout drain).
type lockingWriter struct {
	mu *sync.Mutex
	w  *bytes.Buffer
}

func (lw *lockingWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

var listenURLRe = regexp.MustCompile(`https?://127\.0\.0\.1:\d+/?`)

// extractListenURL finds the first localhost listen URL printed by wrk --web.
func extractListenURL(stdout string) string {
	return listenURLRe.FindString(stdout)
}

// runWebProbe starts long-lived wrk --web, waits for the listen URL on stdout,
// HTTP GETs req.WebPath (default "/"), kills the process, and returns stdout,
// stderr, and HTTPStatus/HTTPBody. Always kills in a deferred path so the suite
// never hangs.
func runWebProbe(t *testing.T, req *Request, bin string, args []string) (*Response, error) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	if req.RepoDir != "" {
		cmd.Dir = req.RepoDir
	} else {
		cmd.Dir = req.WorkRoot
	}
	cmd.Env = wrkEnv(req)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var (
		stdoutMu  sync.Mutex
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	go func() { _, _ = io.Copy(&lockingWriter{mu: &stdoutMu, w: &stdoutBuf}, stdoutPipe) }()
	go func() { _, _ = io.Copy(&stderrBuf, stderrPipe) }()

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	var (
		waitErr error
		waited  bool
		waitMu  sync.Mutex
		once    sync.Once
	)
	markWaited := func(err error) {
		waitMu.Lock()
		defer waitMu.Unlock()
		if waited {
			return
		}
		waited = true
		waitErr = err
	}
	isWaited := func() bool {
		waitMu.Lock()
		defer waitMu.Unlock()
		return waited
	}
	// killOnce terminates the process if still running, then reaps Wait at most once.
	killOnce := func() {
		once.Do(func() {
			if !isWaited() && cmd.Process != nil {
				_ = cmd.Process.Signal(syscall.SIGTERM)
			}
			if isWaited() {
				return
			}
			select {
			case err := <-waitCh:
				markWaited(err)
			case <-time.After(2 * time.Second):
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				markWaited(<-waitCh)
			}
		})
	}
	defer killOnce()

	// Poll stdout for listen URL (timeout ~10s) or early process exit.
	var listenURL string
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && listenURL == "" {
		if isWaited() {
			break
		}
		select {
		case err := <-waitCh:
			markWaited(err)
		default:
		}
		if isWaited() {
			break
		}
		stdoutMu.Lock()
		s := stdoutBuf.String()
		stdoutMu.Unlock()
		if u := extractListenURL(s); u != "" {
			listenURL = u
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	out := &Response{}
	path := req.WebPath
	if path == "" {
		path = "/"
	}
	if listenURL != "" && !isWaited() {
		base := strings.TrimRight(listenURL, "/")
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		getURL := base + path
		client := &http.Client{Timeout: 5 * time.Second}
		if hr, err := client.Get(getURL); err == nil {
			body, _ := io.ReadAll(hr.Body)
			_ = hr.Body.Close()
			out.HTTPStatus = hr.StatusCode
			out.HTTPBody = string(body)
		}
	}

	// Always stop the server before returning final buffers.
	killOnce()
	time.Sleep(50 * time.Millisecond)

	stdoutMu.Lock()
	out.Stdout = stdoutBuf.String()
	stdoutMu.Unlock()
	out.Stderr = stderrBuf.String()
	waitMu.Lock()
	if waited {
		if waitErr != nil {
			if ee, ok := waitErr.(*exec.ExitError); ok {
				out.ExitCode = ee.ExitCode()
			} else {
				// signalled exit still surfaces as ExitError on most platforms
				out.ExitCode = -1
			}
		} else {
			out.ExitCode = 0
		}
	}
	waitMu.Unlock()
	return out, nil
}

func ensureHelpersUsed() {
	_ = mkdirAll
	_ = writeFile
	_ = skipIfNoGit
	_ = runGitIsolated
	_ = gitOutputIsolated
	_ = gitWorktreeListIsolated
	_ = initGitRepoOnMain
	_ = cloneRepoFromSeed
	_ = ensureSeed
	_ = sanitizeBranchToken
	_ = worktreePath
	_ = branchName
	_ = runWrkFrom
	_ = runWrkWithArgs
	_ = assertErrIsNil
	_ = assertFileExists
	_ = assertGitFileIsWorktreeLink
	_ = assertBranchExists
	_ = assertWorktreeListContains
	_ = assertBranchCheckedOutInWorktree
	_ = v2StdoutTemplate
	_ = joinStdoutBlocks
	_ = assertStdoutExactPath
	_ = assertContains
	_ = assertNotContains
	_ = assertFileNotExists
	_ = assertBranchNotExists
	_ = assertWorktreeListNotContains
	_ = setupWrkWorktreeFromMain
	_ = commitAheadOnWorktree
	_ = revParseHEAD
	_ = assertHEAD
	_ = assertEmptyStderr
	_ = syncCommitWord
	_ = syncDetailPass2
	_ = syncSummaryLine
	_ = buildSyncStdout
	_ = primaryThenSyncStdout
	_ = compositionResolvePath
	_ = setupCompositionTwoWTs
	_ = setupConsumerWithAheadExternalDep
	_ = prepareAheadExternalDepConsumerForDone
	_ = runGoModInDir
	_ = buildWrkCLIArgs
	_ = execScriptTTYWrk
	_ = captureCommandOutput
	_ = slugify
	_ = worktreePathWithTask
	_ = branchNameWithTask
	_ = createWorktreeWithTask
	_ = wrkEnv
	_ = appendExtraEnv
	_ = prependPATH
	_ = appendCDEnv
	_ = freeLocalPort
	_ = extractListenURL
	_ = runWebProbe
	_ = lockingWriter{}
}
```