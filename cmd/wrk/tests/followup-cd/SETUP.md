# Scenario

**Feature**: wrk bash follow-up auto-cd (script surface, binary protocol, wrapper)

```
# create: home-gated write when channel open (default location only)
WRK_FOLLOWUP_FILE=tmp; cwd=home; wrk <repo> -> file: cd /abs-worktree
WRK_FOLLOWUP_FILE=tmp; cwd=main; wrk -> file empty
WRK_FOLLOWUP_FILE=tmp; cwd=home; wrk <repo> <target> -> file empty (never)

# done/set-task: existence-gated write when channel open
WRK_FOLLOWUP_FILE=tmp wrk [--done|--set-task] -> file: cd /abs (iff shell cwd gone)

# --force-cd: bypass gates; dual path land
WRK_FOLLOWUP_FILE=tmp; cwd=main; wrk --force-cd -> file: cd /abs-worktree
WRK_FOLLOWUP_FILE unset; fake bash; wrk --force-cd -> install hint; shell @ dest
wrk --force-cd --no-cd -> hard error

# wrapper sources bash.sh, runs binary, executes whitelisted cd
source bash.sh; wrk ... -> stderr "cd /abs"; builtin cd; pwd changes
```

## Preconditions

- The wrk Go module main package is two levels above this tree (`go-pkgs/cmd/wrk/`).
- Go toolchain and git are available on PATH.
- Session-built `wrk` binary at `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`
  (file-locked across leaf processes).
- Each leaf uses isolated `{WorkRoot}/.wrk`, fake `{WorkRoot}/home`, and
  `WRK_DATE=2026-06-30`.
- Tests export `HOME=FakeHome`; `os.UserHomeDir()` resolves to FakeHome on Unix,
  so create home-gate leaves use shell cwd = FakeHome for “at home” cases.

## Steps

1. Root `Setup` allocates WorkRoot / WRK_HOME / FakeHome and registers helpers.
2. Descendants set `req.Mode`, git fixtures, follow-up paths, and CLI args.

## Context

- Follow-up lines: `cd /absolute/path` only; wrapper prints them to stderr then `cd`s.
- User-facing stdout still ends with trailing `\n` where content is non-empty.
- Bash-integration install modes skip `events.jsonl` for pure install/print/complete;
  binary create/done/set-task still append events (not asserted here unless needed).
- Create home gate uses shell process cwd (`RepoDir` binary / `StartDir` wrapper),
  not the source repo path; home success leaves pass main repo as a positional arg.
- Branch B `--force-cd` leaves **must** call `installFakeBash` so
  `LoginInteractive` cannot hang in CI.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"

	"github.com/xhd2015/gitops/git/git_isolated"
)

const wrkDate = "2026-06-30"

const (
	fixtureSeedMainReadme = "main-readme"
	fixtureSeedMainGoMod  = "main-gomod"
)

type seedBuilder func(seedDir string)

func Setup(t *testing.T, req *Request) error {
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	if req.WrkHome == "" {
		req.WrkHome = filepath.Join(workRoot, ".wrk")
	}
	if req.FakeHome == "" {
		req.FakeHome = filepath.Join(workRoot, "home")
	}
	if req.RepoDir == "" {
		req.RepoDir = workRoot
	}
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.FakeHome, 0o755); err != nil {
		return err
	}
	ensureFollowupHelpersUsed()
	return nil
}

func bashShPath(wrkHome string) string {
	return filepath.Join(wrkHome, "integration", "bash.sh")
}

func readFileIfExists(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func minimalCompletionOnlyBashSh() string {
	return `#!/usr/bin/env bash
# pre-seeded completion-only wrk integration (no wrk wrapper)
_wrk() { :; }
complete -o default -F _wrk wrk
`
}

func defaultFollowupPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "followup.txt")
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

func isValidGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
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

func buildSeedMainGoMod(seedDir string) {
	buildSeedMainReadme(seedDir)
	writeFileSeed(filepath.Join(seedDir, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitSeed(seedDir, "add", "go.mod")
	runGitSeed(seedDir, "commit", "-m", "add go.mod")
}

func ensureSeed(t *testing.T, seedID string, build seedBuilder) string {
	t.Helper()
	seedsDir := filepath.Join(fixtureSessionRoot(t), "seeds")
	seedDir := filepath.Join(seedsDir, seedID)
	if isValidGitRepo(seedDir) {
		if resolved, err := filepath.EvalSymlinks(seedDir); err == nil {
			seedDir = resolved
		}
		return seedDir
	}
	lockPath := filepath.Join(seedsDir, ".lock-"+seedID)
	withFlock(t, lockPath, func() {
		if isValidGitRepo(seedDir) {
			return
		}
		_ = os.RemoveAll(seedDir)
		if err := os.MkdirAll(seedDir, 0o755); err != nil {
			t.Fatalf("mkdir seed %s: %v", seedDir, err)
		}
		build(seedDir)
	})
	if resolved, err := filepath.EvalSymlinks(seedDir); err == nil {
		seedDir = resolved
	}
	if !isValidGitRepo(seedDir) {
		t.Fatalf("seed %q not built", seedID)
	}
	return seedDir
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

func runWrkWithArgs(t *testing.T, req *Request, dir string, args ...string) string {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		"HOME=" + req.FakeHome,
		"WRK_HOME=" + req.WrkHome,
		"WRK_DATE=" + wrkDate,
		"PATH=" + os.Getenv("PATH"),
	}
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk %v exit %d stderr=%q", args, ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func setupMainRepo(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
	return mainRepo
}

func setupWrkWorktreeFromMain(t *testing.T, req *Request) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = setupMainRepo(t, req)
	wtDir = runWrkWithArgs(t, req, mainRepo)
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

func assertStdoutEndsWithNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		return
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; got %q", stdout)
	}
}

func assertFollowupEmpty(t *testing.T, resp *Response) {
	t.Helper()
	if !resp.FollowupExists {
		return
	}
	if strings.TrimSpace(resp.FollowupContent) != "" {
		t.Fatalf("expected empty follow-up file, got %q", resp.FollowupContent)
	}
}

func assertFollowupCD(t *testing.T, resp *Response, wantAbs string) {
	t.Helper()
	if !resp.FollowupExists {
		t.Fatalf("expected follow-up file at %s to exist", resp.FollowupPath)
	}
	wantAbs = resolvePath(t, wantAbs)
	wantLine := "cd " + wantAbs
	got := strings.TrimSpace(resp.FollowupContent)
	if got != wantLine {
		t.Fatalf("follow-up content: want %q, got %q (raw %q)", wantLine, got, resp.FollowupContent)
	}
	// Prefer trailing newline after the line (shell-friendly).
	if resp.FollowupContent != "" && !strings.HasSuffix(resp.FollowupContent, "\n") {
		t.Fatalf("follow-up file should end with newline; got %q", resp.FollowupContent)
	}
}

// installFakeBash places a non-interactive bash shim first on PATH so
// LoginInteractive (Branch B --force-cd) cannot hang. Call from every
// successful Branch B leaf (and failure leaves that must prove no shell).
func installFakeBash(t *testing.T, req *Request, exitCode int) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "fake-shell-bin")
	mkdirAll(t, binDir)
	logPath := filepath.Join(req.WorkRoot, "fake-shell-log.txt")
	writeFile(t, logPath, "")
	fake := filepath.Join(binDir, "bash")
	body := `#!/bin/sh
log="${WRK_FAKE_SHELL_LOG:-}"
if [ -n "$log" ]; then
  printf 'cwd=%s\n' "$(pwd)" >> "$log"
  printf 'args=%s\n' "$*" >> "$log"
fi
code="${WRK_FAKE_SHELL_EXIT:-0}"
exit "$code"
`
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake bash: %v", err)
	}
	req.FakeShellDir = binDir
	req.FakeShellLog = logPath
	req.FakeShellExit = exitCode
	req.ShellEnv = fake
}

func assertFakeShellCwd(t *testing.T, req *Request, wantAbs string) {
	t.Helper()
	if req.FakeShellLog == "" {
		t.Fatal("FakeShellLog not set")
	}
	data, err := os.ReadFile(req.FakeShellLog)
	if err != nil {
		t.Fatalf("read fake shell log: %v", err)
	}
	wantAbs = resolvePath(t, wantAbs)
	wantLine := "cwd=" + wantAbs
	if !strings.Contains(string(data), wantLine) {
		t.Fatalf("fake shell log missing %q; full log:\n%s", wantLine, data)
	}
}

func assertFakeShellLaunched(t *testing.T, req *Request) {
	t.Helper()
	if req.FakeShellLog == "" {
		t.Fatal("FakeShellLog not set")
	}
	data, err := os.ReadFile(req.FakeShellLog)
	if err != nil {
		t.Fatalf("read fake shell log: %v", err)
	}
	if !strings.Contains(string(data), "cwd=") {
		t.Fatalf("fake shell was not launched; log=%q", data)
	}
}

func assertFakeShellNotLaunched(t *testing.T, req *Request) {
	t.Helper()
	if req.FakeShellLog == "" {
		t.Fatal("FakeShellLog not set (installFakeBash required to detect accidental launch)")
	}
	data, err := os.ReadFile(req.FakeShellLog)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read fake shell log: %v", err)
	}
	if strings.Contains(string(data), "cwd=") {
		t.Fatalf("fake shell should not have launched; log=%q", data)
	}
}

func assertInstallHint(t *testing.T, stderr string) {
	t.Helper()
	assertContains(t, stderr, "wrk --bash-integration --install")
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	}
}

func assertPathsEqual(t *testing.T, got, want string) {
	t.Helper()
	g := resolvePath(t, got)
	w := resolvePath(t, want)
	if g != w {
		t.Fatalf("path: want %q, got %q", w, g)
	}
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	if path == "" {
		return ""
	}
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

func requireMode(t *testing.T, req *Request, mode string) {
	t.Helper()
	if req.Mode != mode {
		t.Fatalf("expected mode %q, got %q", mode, req.Mode)
	}
}

func ensureFollowupHelpersUsed() {
	_ = bashShPath
	_ = readFileIfExists
	_ = minimalCompletionOnlyBashSh
	_ = defaultFollowupPath
	_ = skipIfNoGit
	_ = mkdirAll
	_ = writeFile
	_ = runGitIsolated
	_ = gitOutputIsolated
	_ = initGitRepoOnMain
	_ = cloneRepoFromSeed
	_ = ensureSeed
	_ = worktreePath
	_ = branchName
	_ = runWrkWithArgs
	_ = setupMainRepo
	_ = setupWrkWorktreeFromMain
	_ = commitAheadOnWorktree
	_ = slugify
	_ = worktreePathWithTask
	_ = branchNameWithTask
	_ = assertContains
	_ = assertNotContains
	_ = assertStdoutEndsWithNewline
	_ = assertFollowupEmpty
	_ = assertFollowupCD
	_ = installFakeBash
	_ = assertFakeShellCwd
	_ = assertFakeShellLaunched
	_ = assertFakeShellNotLaunched
	_ = assertInstallHint
	_ = assertFileExists
	_ = assertFileNotExists
	_ = assertPathsEqual
	_ = resolvePath
	_ = requireMode
}
```
