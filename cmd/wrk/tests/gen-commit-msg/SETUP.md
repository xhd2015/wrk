# Scenario

**Feature**: wrk --gen-commit-msg wires agent-pro commit_msg.RunGenCommitMsg

```
# mode + flag forwarding (end-to-end through wrk binary)
wrk --gen-commit-msg [flags...]
  -> commit_msg.RunGenCommitMsg(remaining flags)

# help
wrk --gen-commit-msg -h | --help
  -> documents gen-commit-msg options (model, dry-run, commit, no-verify, agent-runner, add-all)

# dry-run pure plan (library mock B)
git repo with staged files
  -> wrk --gen-commit-msg --dry-run [--model …]
  -> stdout: dry-run: would generate commit message for N staged file(s)\n
  -> exit 0; no agent
  -> binary staged: would-unstage on stderr; index unchanged
  -> --commit / --commit --no-verify: would: git commit on stderr; HEAD unchanged
  -> --add-all: would: git add -A on stderr; no unrecognized flag

# compose peel (Classic RED until genCommitMsgBoolFlags includes --add-all)
linked wt + staged
  -> wrk --gen-commit-msg --add-all --commit --done --dry-run
  -> not: unrecognized flag: --add-all
  -> stderr: would: git add -A; gen-commit + primary dry plan; exit 0

# generate / commit via fake-opencode (no live LLM)
staged + FAKE_OPENCODE_MOCK_CONFIG + --agent-runner-binary <fake-opencode>
  -> wrk --gen-commit-msg --agent-runner opencode --model openai/gpt-5
  -> stdout: parsed title + description
  -> optional --commit / --commit --no-verify

# validation / mutex vs pipeline compose
wrk --gen-commit-msg --status              -> mutually exclusive
wrk --gen-commit-msg --sync [--dry-run]     -> allowed multi-stage (activeRoot pipeline)
wrk --gen-commit-msg --no-verify           -> requires --commit
wrk --gen-commit-msg --dry-run --agent-runner codex
  -> unsupported agent runner

# P2 done/merge-back compose is under monotree done-compose/ + done-pipeline/dry-run/
# (not this nested root): --gen-commit-msg --commit --done|…
```

## Preconditions

- The wrk Go module root is three levels above this tree root
  (`cmd/wrk/tests/gen-commit-msg` → module at workspace root).
- Go toolchain is available on PATH.
- Session wrk binary is built once per doctest run to
  process-local `MkdirTemp` wrk binary (in-memory mutex).
- Session fake-opencode is built once to
  process-local fake-opencode binary from
  `external/agent-pro-master-2026-07-16/cmd/fake-opencode`.
- Git is required for dry-run / generate / commit success leaves
  (and unknown-agent-runner which stages a file).
- Coverage backfill: production wire (wrkcli → commit_msg) is already present → GREEN.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Default `RepoDir` is a neutral empty workspace (no git) unless a leaf stages a repo.
3. Leaves set `req.Args` and optionally init a git repo with staged files.
4. fake-opencode leaves write mock config, call `installFakeOpencodeEnv`, set agent args.

## Context

- Dry-run leaves use hooks-disabled isolated git (`core.hooksPath=/dev/null`).
- Success dry-run leaves assert exact mock message B via v2 full-match stdout templates.
- Error leaves assert non-zero exit + stderr substrings (`mutually exclusive`,
  `--no-verify`/`--commit`, unsupported agent runner / `codex`).
- Pure dry-run: `WRK_HOME` isolation only; no FAKE_OPENCODE / agent binary.
- generate/commit: ExtraEnv carries `FAKE_OPENCODE_MOCK_CONFIG` + `OPENCODE_CONFIG_DIR`;
  Args include `--agent-runner-binary` pointing at session fake-opencode.

```go
import (
	"bytes"
	"github.com/xhd2015/wrk/wrkcli"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
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

// Process-local fake-opencode binary (one-process suite; in-memory mutex).
var (
	fakeOpencodeBinMu   sync.Mutex
	fakeOpencodeBinPath string
	fakeOpencodeBinErr  error
)

func getFakeOpencodeBin(t *testing.T) string {
	t.Helper()
	fakeOpencodeBinMu.Lock()
	defer fakeOpencodeBinMu.Unlock()
	if fakeOpencodeBinPath != "" || fakeOpencodeBinErr != nil {
		if fakeOpencodeBinErr != nil {
			t.Fatal(fakeOpencodeBinErr)
		}
		return fakeOpencodeBinPath
	}
	dir, err := os.MkdirTemp("", "wrk-doctest-fake-opencode-")
	if err != nil {
		fakeOpencodeBinErr = err
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "fake-opencode")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/fake-opencode")
	cmd.Dir = agentProRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fakeOpencodeBinErr = fmt.Errorf("build fake-opencode: %v\n%s", err, out)
		t.Fatal(fakeOpencodeBinErr)
	}
	fakeOpencodeBinPath = bin
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
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	ensureGenCommitMsgHelpersUsed()
	return nil
}

func genCommitMsgWrkEnv(req *Request) []string {
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
}

func agentProRoot(t *testing.T) string {
	t.Helper()
	modRoot := wrkModRoot
	if modRoot == "" {
		t.Fatal("wrkModRoot unset; root Setup must call noteModRoot first")
	}
	root := filepath.Join(modRoot, "external", "agent-pro-master-2026-07-16")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		// CI checkouts do not vendor this external tree; skip integration leaves.
		t.Skipf("agent-pro root not found at %s: %v", root, err)
	}
	return root
}

// writeMockConfig writes FAKE_OPENCODE_MOCK_CONFIG JSON under WorkRoot.
func writeMockConfig(t *testing.T, req *Request, body string) {
	t.Helper()
	req.MockConfigPath = filepath.Join(req.WorkRoot, "fake-opencode-mock.json")
	writeFile(t, req.MockConfigPath, body)
}

// installFakeOpencodeEnv builds fake-opencode (session-cached) and sets ExtraEnv
// for FAKE_OPENCODE_MOCK_CONFIG + OPENCODE_CONFIG_DIR. Requires MockConfigPath.
func installFakeOpencodeEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.MockConfigPath == "" {
		t.Fatal("installFakeOpencodeEnv: MockConfigPath is empty; call writeMockConfig first")
	}
	if req.FakeOpencode == "" {
		req.FakeOpencode = getFakeOpencodeBin(t)
	}
	opencodeConfigDir := filepath.Join(req.WorkRoot, "opencode-config")
	if err := os.MkdirAll(opencodeConfigDir, 0o755); err != nil {
		t.Fatalf("mkdir OPENCODE_CONFIG_DIR: %v", err)
	}
	req.ExtraEnv = append(req.ExtraEnv,
		"FAKE_OPENCODE_MOCK_CONFIG="+req.MockConfigPath,
		"OPENCODE_CONFIG_DIR="+opencodeConfigDir,
	)
}

// genCommitMsgAgentArgs builds wrk args for the fake-opencode agent path.
// extra is appended after the common agent-runner flags (e.g. --commit).
// Requires installFakeOpencodeEnv first (sets FakeOpencode).
func genCommitMsgAgentArgs(req *Request, extra ...string) []string {
	args := []string{
		"--gen-commit-msg",
		"--agent-runner", "opencode",
		"--agent-runner-binary", req.FakeOpencode,
		"--model", "openai/gpt-5",
	}
	return append(args, extra...)
}

const mockConfigAddFeature = `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_commit","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: add feature\", \"description\": \"Implement feature X\"}"},{"type":"step_finish"}]}`

const mockConfigSkipHooks = `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_no_verify","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: skip hooks\", \"description\": \"Commit with --no-verify\"}"},{"type":"step_finish"}]}`

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
	// Isolate from user/global config; leave room for repo-local hooksPath.
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

// initGitRepoWithFailingPreCommitHook creates a repo with a pre-commit hook that exits 1.
// Initial commit uses hooks disabled; then hooksPath is pointed at .git/hooks with a failing pre-commit.
func initGitRepoWithFailingPreCommitHook(t *testing.T, dir string) {
	t.Helper()
	skipIfNoGit(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	// Disable hooks for the seed commit only.
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
	// Re-enable real hooks so pre-commit runs (unless --no-verify).
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

// stageBinaryAndTextFile stages app.go + blob.bin (ELF magic). Returns binary rel path.
func stageBinaryAndTextFile(t *testing.T, req *Request) string {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepo(t, repo)
	writeFile(t, filepath.Join(repo, "app.go"), "package main\n")
	runGit(t, repo, "add", "app.go")

	binaryRel := "blob.bin"
	binPath := filepath.Join(repo, binaryRel)
	// Minimal ELF magic so detect.DetectFileType treats it as binary.
	if err := os.WriteFile(binPath, []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01}, 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	runGit(t, repo, "add", binaryRel)
	req.RepoDir = repo
	req.BinaryRel = binaryRel
	return binaryRel
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

func mockMessageB(n int) string {
	return fmt.Sprintf("dry-run: would generate commit message for %d staged file(s)\n", n)
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

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
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

func assertMockMessageB(t *testing.T, stdout string, n int) {
	t.Helper()
	assertOutputExact(t, stdout, v2StdoutTemplate(mockMessageB(n)))
}

func ensureGenCommitMsgHelpersUsed() {
	_ = stageOneTextFile
	_ = stageBinaryAndTextFile
	_ = stageOneTextFileWithFailingPreCommit
	_ = initGitRepoWithFailingPreCommitHook
	_ = gitHEADSubject
	_ = gitStagedNames
	_ = mockMessageB
	_ = assertMockMessageB
	_ = assertExitZero
	_ = assertExitNonZero
	_ = assertOutputExact
	_ = skipIfNoGit
	_ = getFakeOpencodeBin
	_ = writeMockConfig
	_ = installFakeOpencodeEnv
	_ = genCommitMsgAgentArgs
}


// runCLIWithEnv runs the product binary with cmd.Dir/cmd.Env (Parallel-safe).
// Named historically; DOCTEST_LINT.md §1 forbids Setenv/Chdir for leaf isolation.
func runCLIWithEnv(t *testing.T, dir, wrkHome string, args, env []string) (*Response, error) {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = env
	} else {
		cmd.Env = append(os.Environ(), "WRK_HOME="+wrkHome)
	}
	return captureCommandOutput(cmd, "")
}

func captureCommandOutput(cmd *exec.Cmd, stdinInput string) (*Response, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdinInput != "" {
		cmd.Stdin = strings.NewReader(stdinInput)
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


```
