# Scenario

**Feature**: wrk CLI auto worktree, merge-back, and worktree listing

```
# isolated WRK_HOME + work root per test; build wrk once
wrk (no args) from cwd -> stdout path only + git worktree side effects
wrk --done [--confirm-from-stdin] from linked wt -> merge-back --rm
wrk --list from cwd -> git worktree list stdout unchanged
```

## Preconditions

- The wrk Go module root is three levels above the test tree root (at `cmd/wrk/tests/`)
- Go toolchain is available on PATH
- Git is required for worktree tests

## Context

Each test runs the `wrk` CLI in an isolated environment. The `wrk` binary is built once per doctest session (`{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`, file-locked across leaf processes). Each leaf gets its own temp directory and isolated `WRK_HOME` at `{WorkRoot}/.wrk`.

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
	"github.com/xhd2015/gitops/git/git_isolated"
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
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, req *Request) error {
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
	return runWrkWithArgs(t, req, dir)
}

func runWrkWithArgs(t *testing.T, req *Request, dir string, args ...string) string {
	t.Helper()
	bin := getWrkBin(t)
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
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
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

	writeFile(t, filepath.Join(externalPath, "dep.go"), "package dep // ahead fix\n")
	runGitIsolated(t, externalPath, "add", "dep.go")
	runGitIsolated(t, externalPath, "commit", "-m", "dep fix on external worktree")
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
		env := []string{
			"HOME=" + home,
			"PATH=/usr/bin:/bin",
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
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
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
	_ = setupConsumerWithAheadExternalDep
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
	_ = prepareFollowupFile
	_ = freeLocalPort
	_ = extractListenURL
	_ = runWebProbe
	_ = lockingWriter{}
}
```