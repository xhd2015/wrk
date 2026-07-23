# Scenario

**Feature**: wrk --tag-next plans and applies per-scope release tags

```
# isolated WRK_HOME + git repo per test; build wrk once per session
git repo + version tags -> wrk --tag-next [--dry-run] [--push] [--json] -> stdout plan/apply summary
```

## Preconditions

- The wrk Go module root is three levels above this test tree (`cmd/wrk/tests/tag-next/`).
- Go toolchain and git are available on PATH.
- Process-local wrk binary built once under an in-memory mutex (not session flock).
- Git helpers use `github.com/xhd2015/gitops/git/git_isolated` (hook-free).

## Context

- Each leaf gets isolated `{WorkRoot}` and `{WorkRoot}/.wrk` as `WRK_HOME`.
- `initTaggedRepo` builds repos with lightweight tags at explicit commits.
- `runWrkTagNext` is satisfied by root `Run` + `req.Args`.
- Stdout assertions use `assert.Output` v2 templates where output is bounded.
- Dynamic short hashes use regex lines or runtime `git rev-parse` in `Assert`.

```go
import (
	"bytes"
	"github.com/xhd2015/wrk/wrkcli"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	tagNextEnsureHelpersUsed()
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

func createLightweightTag(t *testing.T, repo, name, ref string) {
	t.Helper()
	if ref == "" {
		ref = "HEAD"
	}
	runGitIsolated(t, repo, "tag", name, ref)
}

func shortHEAD(t *testing.T, repo string) string {
	t.Helper()
	return gitOutputIsolated(t, repo, "rev-parse", "--short=7", "HEAD")
}

func tagRefExists(t *testing.T, repo, name string) bool {
	t.Helper()
	err := git_isolated.Command(repo, "rev-parse", "--verify", "refs/tags/"+name).Run()
	return err == nil
}

func remoteTagExists(t *testing.T, bareOrigin, name string) bool {
	t.Helper()
	out := gitOutputIsolated(t, bareOrigin, "show-ref", "--tags")
	prefix := "refs/tags/" + name
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == prefix {
			return true
		}
	}
	return false
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

func setupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func attachOriginAndPushMain(t *testing.T, repo, bareOrigin string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "add", "origin", bareOrigin)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
}

// setupRootBumpRepo: v0.0.1 at first commit, second commit changes README.
func setupRootBumpRepo(t *testing.T, req *Request) string {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# v1\n", "init")
	createLightweightTag(t, repo, "v0.0.1", "")
	commitFile(t, repo, "README.md", "# v2\n", "bump root")
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	return repo
}

// setupNoChangeRepo: tag v0.0.1 at HEAD; no post-tag commits.
func setupNoChangeRepo(t *testing.T, req *Request) string {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# stable\n", "init")
	createLightweightTag(t, repo, "v0.0.1", "")
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	return repo
}

// setupPrereleaseSkipRepo: releases v0.0.1/v0.0.2, then v0.0.3-alpha at HEAD.
func setupPrereleaseSkipRepo(t *testing.T, req *Request) string {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# base\n", "init")
	createLightweightTag(t, repo, "v0.0.1", "")
	createLightweightTag(t, repo, "v0.0.2", "")
	head := commitFile(t, repo, "README.md", "# changed\n", "post-release change")
	createLightweightTag(t, repo, "v0.0.3-alpha", head)
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	return repo
}

// setupSubScopeOnlyRepo: root v0.0.1 + sub/v0.2.3 at baseline; only sub/ changes.
func setupSubScopeOnlyRepo(t *testing.T, req *Request) string {
	t.Helper()
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	writeFile(t, filepath.Join(repo, "README.md"), "# root\n")
	writeFile(t, filepath.Join(repo, "sub", "lib.go"), "package sub\n")
	runGitIsolated(t, repo, "add", "README.md", "sub/lib.go")
	runGitIsolated(t, repo, "commit", "-m", "init root and sub")
	createLightweightTag(t, repo, "v0.0.1", "")
	createLightweightTag(t, repo, "sub/v0.2.3", "")
	commitFile(t, repo, "sub/lib.go", "package sub // changed\n", "change sub only")
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	return repo
}

// setupPushRepo: root bump repo with bare origin remote.
func setupPushRepo(t *testing.T, req *Request) string {
	t.Helper()
	bare := setupBareOrigin(t, req.WorkRoot, "origin")
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# v1\n", "init")
	createLightweightTag(t, repo, "v0.0.1", "")
	commitFile(t, repo, "README.md", "# v2\n", "bump root")
	attachOriginAndPushMain(t, repo, bare)
	repo = resolveRepoPath(t, repo)
	req.MainRepo = repo
	req.RepoDir = repo
	req.OriginBare = bare
	return repo
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	mkdirAll(t, cwd)
	return cwd
}

func eventsJSONLPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

type wrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func readEvents(t *testing.T, wrkHome string) []wrkEvent {
	t.Helper()
	data, err := os.ReadFile(eventsJSONLPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []wrkEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev wrkEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
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

func tagNextStdoutV2(body string) string {
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
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if regexp.MustCompile(`\x1b\[[0-9;]*m`).MatchString(s) {
		t.Fatalf("stdout should not contain ANSI escapes, got %q", s)
	}
}

func assertValidJSONObject(t *testing.T, s string) map[string]any {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &obj); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, s)
	}
	return obj
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func buildTagNextCLIArgs(req *Request) []string {
	var args []string
	if req.TargetDir != "" {
		args = append(args, req.TargetDir)
	}
	args = append(args, req.Args...)
	return args
}

func tagNextWrkEnv(req *Request) []string {
	return append(os.Environ(), "WRK_HOME="+req.WrkHome)
}

func tagNextEnsureHelpersUsed() {
	_ = buildTagNextCLIArgs
	_ = tagNextWrkEnv
	_ = setupRootBumpRepo
	_ = setupNoChangeRepo
	_ = setupPrereleaseSkipRepo
	_ = setupSubScopeOnlyRepo
	_ = setupPushRepo
	_ = shortHEAD
	_ = tagRefExists
	_ = remoteTagExists
	_ = readEvents
	_ = tagNextStdoutV2
	_ = assertNoANSI
	_ = assertValidJSONObject
	_ = initNeutralCwd
}

// runCLIWithEnv maps a full env slice (as historically used with exec.Command)
// onto wrkcli.RunOptions for L2 in-process CLI.
func runCLIWithEnv(t *testing.T, dir, wrkHome string, args, env []string) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	opts := wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     dir,
		WrkHome: wrkHome,
	}
	for _, e := range env {
		key, val, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		switch key {
		case "WRK_HOME":
			if strings.TrimSpace(val) != "" {
				opts.WrkHome = val
			}
		case "WRK_DATE":
			opts.WrkDate = val
		default:
			opts.ExtraEnv = append(opts.ExtraEnv, e)
		}
	}
	code := wrkcli.RunCLI(args, opts)
	return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}, nil
}

```
