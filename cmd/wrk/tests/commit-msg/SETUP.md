# Scenario

**Feature**: wrk --commit -m/--message supplies a user commit message (no AI)

```
# validation (early, non-zero, stderr intent-stable)
wrk -m "x"                         -> requires --commit
wrk --message "x"                  -> requires --commit
wrk --gen-commit-msg --commit -m x -> mutually exclusive / exclusive
wrk --commit                       -> needs -m/--message or --gen-commit-msg
wrk --commit -m "" / "   "         -> empty / invalid message

# apply (wrk-owned git commit stage)
staged -> wrk --commit -m "feat: x" [--no-verify] [--add-all]
  -> HEAD subject = first line of message
  -> --dry-run: would: git commit + message; HEAD unchanged

# shared branch
two checkouts of same branch + staged
  -> wrk --commit -m "x" -> Error: refuse commit

# compose (same partners as gen-commit-msg)
wrk --commit -m "x" --done [--dry-run]
  -> must NOT stderr "mutually exclusive"
  -> may later fail (not a linked worktree / nothing staged)
# clean + compose: soft-skip only when -m already matches HEAD
wrk --commit -m "initial" --exec true   -> notice: worktree clean, skip commit
wrk --commit -m "feat: other" --exec true -> still nothing to commit

# help
wrk -h -> documents -m/--message, requires --commit, exclusive with gen-commit-msg
```

## Preconditions

- The wrk Go module root is three levels above this tree root
  (`cmd/wrk/tests/commit-msg` → module at workspace root).
- Go toolchain is available on PATH.
- Session wrk binary is built once per doctest run to
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk` (used when `InProcess`
  is false; default leaves prefer `InProcess` Capture).
- Git is required for apply / shared-branch / compose fixture leaves.
- Classic TDD **RED** until product implements manual message flags and stage.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Default `RepoDir` is a neutral empty workspace (no git) unless a leaf stages a repo.
3. Leaves set `req.Args`, optionally init git fixtures, set `req.InProcess = true`.

## Context

- Prefer **L2** `wrkcli.Capture` (`InProcess`) for all leaves in this tree.
- Error asserts are intent-stable: mention flags + require/exclusive/empty;
  product wording may be `wrk: …` or `Error: …`.
- Parallel-safe: inject Dir/Env on Capture or child process; never `t.Setenv` /
  `t.Chdir` / `os.Chdir` for product cwd.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// harnessDoctest holds inject fields from d (no os.Setenv — Parallel-safe).
var (
	harnessMu          sync.Mutex
	harnessSessionID   string
	harnessDoctestRoot string
)

func adoptDoctestContext(d *session.Doctest) {
	if d == nil {
		return
	}
	harnessMu.Lock()
	defer harnessMu.Unlock()
	if d.DOCTEST_SESSION_ID != "" {
		harnessSessionID = d.DOCTEST_SESSION_ID
	}
	if d.DOCTEST_ROOT != "" {
		harnessDoctestRoot = d.DOCTEST_ROOT
	}
}

func doctestSessionID(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	sid := harnessSessionID
	harnessMu.Unlock()
	// Parallel-safe: only d.DOCTEST_SESSION_ID via adoptDoctestContext (no os.Getenv).
	if sid == "" {
		t.Fatal("d.DOCTEST_SESSION_ID not set (expected adoptDoctestContext from Setup)")
	}
	return sid
}

func doctestRootPath(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	root := harnessDoctestRoot
	harnessMu.Unlock()
	// Parallel-safe: only d.DOCTEST_ROOT via adoptDoctestContext (no os.Getenv).
	if root == "" {
		t.Fatal("d.DOCTEST_ROOT not set (expected adoptDoctestContext from Setup)")
	}
	return root
}

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
	return filepath.Join(fixtureCacheBase(t), doctestSessionID(t))
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
		modRoot := findModuleRoot(doctestRootPath(t))
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
		}
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			t.Fatalf("mkdir bin dir: %v", err)
		}
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	adoptDoctestContext(d)
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	ensureCommitMsgHelpersUsed()
	return nil
}

func commitMsgWrkEnv(req *Request) []string {
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", cwd, err)
	}
	return cwd
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s in %s failed: %s: %v", strings.Join(args, " "), dir, string(out), err)
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

// initGitRepo creates a hooks-disabled git repo with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	skipIfNoGit(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	writeFile(t, filepath.Join(dir, "README.md"), "initial\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
}

// stageOneTextFile inits a repo under WorkRoot/repo, stages change.go, sets RepoDir.
func stageOneTextFile(t *testing.T, req *Request) {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepo(t, repo)
	writeFile(t, filepath.Join(repo, "change.go"), "package main\n")
	runGit(t, repo, "add", "change.go")
	req.RepoDir = repo
}

// initCleanGitRepo inits hooks-disabled repo with only the seed commit (nothing staged).
func initCleanGitRepo(t *testing.T, req *Request) {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepo(t, repo)
	req.RepoDir = repo
}

// placeUntrackedTextFile inits repo and writes untracked change.go (not staged).
func placeUntrackedTextFile(t *testing.T, req *Request) {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepo(t, repo)
	writeFile(t, filepath.Join(repo, "change.go"), "package main\n")
	req.RepoDir = repo
}

// initGitRepoWithFailingPreCommitHook creates a repo with a pre-commit hook that exits 1.
func initGitRepoWithFailingPreCommitHook(t *testing.T, dir string) {
	t.Helper()
	skipIfNoGit(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	writeFile(t, filepath.Join(dir, "README.md"), "initial\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks dir: %v", err)
	}
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write pre-commit hook: %v", err)
	}
	runGit(t, dir, "config", "core.hooksPath", hooksDir)
}

// stageOneTextFileWithFailingPreCommit stages change.go in a repo with a failing pre-commit.
func stageOneTextFileWithFailingPreCommit(t *testing.T, req *Request) {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepoWithFailingPreCommitHook(t, repo)
	writeFile(t, filepath.Join(repo, "change.go"), "package main\n")
	runGit(t, repo, "add", "change.go")
	req.RepoDir = repo
}

// initGitRepoOnMain creates a bare main-repo checkout under WorkRoot for flag-layer leaves.
func initGitRepoOnMain(t *testing.T, dir string) {
	t.Helper()
	initGitRepo(t, dir)
}

// setupSharedTwoLinkedStaged: main + two linked checkouts of same branch; stage change.go on wt1.
func setupSharedTwoLinkedStaged(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "main")
	initGitRepo(t, mainRepo)
	// Create a feature branch and two worktrees on it.
	runGit(t, mainRepo, "checkout", "-b", "feature-shared")
	wt1 := filepath.Join(req.WorkRoot, "wt1")
	wt2 := filepath.Join(req.WorkRoot, "wt2")
	// Move main back to main so both linked wts hold feature-shared uniquely as pair.
	runGit(t, mainRepo, "checkout", "main")
	runGit(t, mainRepo, "worktree", "add", wt1, "feature-shared")
	runGit(t, mainRepo, "worktree", "add", "--force", wt2, "feature-shared")

	req.MainRepo = mainRepo
	req.WtDir = wt1
	req.Wt2Dir = wt2
	req.WtBranch = "feature-shared"

	writeFile(t, filepath.Join(wt1, "change.go"), "package main\n")
	runGit(t, wt1, "add", "change.go")
	req.RepoDir = wt1
}

func gitHEADSubject(t *testing.T, gitDir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = gitDir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log -1 --format=%%s: %v\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func gitStagedNames(t *testing.T, gitDir string) []string {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = gitDir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git diff --cached --name-only: %v\n%s", err, string(out))
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}

func assertExitNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}

// assertSharedBranchRefuseError checks hard refuse framing for shared-branch commit.
func assertSharedBranchRefuseError(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for shared-branch refuse; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("stderr must include %q; stderr=%q stdout=%q", "Error:", stderr, resp.Stdout)
	}
	if req.WtBranch != "" && !strings.Contains(stderr, req.WtBranch) {
		t.Fatalf("stderr must mention branch %q; stderr=%q", req.WtBranch, stderr)
	}
	if req.WtDir != "" {
		if !strings.Contains(stderr, req.WtDir) && !strings.Contains(stderr, filepath.Base(req.WtDir)) {
			t.Fatalf("stderr must mention primary worktree path %q; stderr=%q", req.WtDir, stderr)
		}
	}
	if req.Wt2Dir != "" {
		if !strings.Contains(stderr, req.Wt2Dir) && !strings.Contains(stderr, filepath.Base(req.Wt2Dir)) {
			t.Fatalf("stderr must mention second worktree path %q; stderr=%q", req.Wt2Dir, stderr)
		}
	}
	low := strings.ToLower(stderr)
	if !strings.Contains(low, "refuse") {
		t.Fatalf("stderr must include refuse language; stderr=%q", stderr)
	}
	if !strings.Contains(low, "commit") {
		t.Fatalf("stderr must name commit op; stderr=%q", stderr)
	}
}

func ensureCommitMsgHelpersUsed() {
	_ = stageOneTextFile
	_ = stageOneTextFileWithFailingPreCommit
	_ = initGitRepoWithFailingPreCommitHook
	_ = initCleanGitRepo
	_ = placeUntrackedTextFile
	_ = initGitRepoOnMain
	_ = setupSharedTwoLinkedStaged
	_ = gitHEADSubject
	_ = gitStagedNames
	_ = assertExitZero
	_ = assertExitNonZero
	_ = assertSharedBranchRefuseError
	_ = skipIfNoGit
}
```
