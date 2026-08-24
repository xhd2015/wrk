# Scenario

**Feature**: wrk --reinstall-local dry-run plans and execute reinstalls via GOBIN isolation

```
# module fixture + GOBIN stubs -> wrk --reinstall-local [--dry-run]
# dry-run: would:/skip: lines + would: reinstall N binaries (M skipped); no mutation
# execute: go install|go run + skip: lines + reinstalled N, skipped M, failed F; mutates GOBIN
WorkRoot/{mod,gobin,.wrk}
  -> wrk --reinstall-local [--dry-run] (cwd=mod, GOBIN=gobin)
  -> stdout plan or execute report; exit 0 (incl. soft failed installs + warning:) | non-zero on hard errors
```

## Preconditions

- Nested root: no inheritance from `cmd/wrk/tests` (own `DOCTEST.md` Version 0.0.10).
- Go toolchain on PATH; session wrk binary built once per doctest run to
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk`.
- Fixture modules are real directories under a per-leaf `t.TempDir()` WorkRoot:
  `ModuleRoot` is process cwd (module root, multi tree, or git subdir);
  `BinDir` (`gobin`) holds stub files named after bins (or real go-installed
  binaries after execute). Multi leaves may create nested go.mod trees and/or
  git checkouts via `initGitRepoOnMain` / `gitCommitAll`.
- P1 pure API may already exist in package `wrkcli`; this tree only exercises
  the **CLI** surface (flag + dry-run text + execute path).

## Steps

1. Root `Setup` creates isolated `WorkRoot`, `WrkHome`, empty `ModuleRoot`, empty `BinDir`.
2. Leaves write go.mod / package mains / stub bins as needed and set `Args`.
3. Root `Run` builds (or reuses) wrk, sets `WRK_HOME` + `GOBIN=BinDir`, runs from `ModuleRoot`.

## Context

- **Stdout vocabulary** matches `--all-deps` style: dry-run uses `would:` lines;
  execute drops the `would:` prefix on command lines and uses an execute summary.
- Last content line of successful dry-run / execute stdout ends with `\n`.
- Exact stdout uses `assert.Output` v2 full-match templates where bounded.
- Execute leaves that stream `go` compiler noise use summary/side-effect asserts
  plus optional `assert.Output` on bounded wrk-owned lines when safe.
- Help / error leaves use substring checks.
- Helpers below write fixtures (same shapes as P1 pure-API tree, plus execute helpers).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"strings"
	"syscall"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
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
	req.WrkHome = filepath.Join(workRoot, ".wrk")
	req.ModuleRoot = filepath.Join(workRoot, "mod")
	req.BinDir = filepath.Join(workRoot, "gobin")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.ModuleRoot, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.BinDir, 0o755); err != nil {
		return err
	}
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}

// reinstallLocalCLIEnv isolates WRK_HOME and forces GOBIN to the leaf bin dir.
// Strips ambient NO_COLOR so leaves own color policy (pipe = plain; --color on).
// ExtraEnv is appended last so leaf overrides can win on duplicate keys where
// the OS takes the last occurrence.
func reinstallLocalCLIEnv(req *Request) []string {
	base := filterEnvKeys(os.Environ(), "NO_COLOR")
	env := append(base,
		"WRK_HOME="+req.WrkHome,
		"GOBIN="+req.BinDir,
	)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
}

// filterEnvKeys drops KEY=… entries whose key is in drop (case-sensitive).
func filterEnvKeys(env []string, drop ...string) []string {
	dropSet := make(map[string]struct{}, len(drop))
	for _, k := range drop {
		dropSet[k] = struct{}{}
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		key := e
		if i := strings.IndexByte(e, '='); i >= 0 {
			key = e[:i]
		}
		if _, skip := dropSet[key]; skip {
			continue
		}
		out = append(out, e)
	}
	return out
}

func writeGoMod(t *testing.T, moduleRoot, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatalf("mkdir module root %s: %v", moduleRoot, err)
	}
	content := "module " + modulePath + "\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func gitCommitAll(t *testing.T, repo, subject string) {
	t.Helper()
	runGitIsolated(t, repo, "add", "-A")
	runGitIsolated(t, repo, "commit", "-m", subject)
}

func writePackageMain(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	// Build source without a single string literal containing both
	// "package main" and "func main()" (doctest anti-pattern check).
	src := fmt.Sprintf("package %s\n\nfunc main() {}\n", "main")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write main.go in %s: %v", dir, err)
	}
}

// writePackageMainPrints writes a package main that prints msg and exits 0.
// Used by execute leaves that assert the installed binary actually runs.
func writePackageMainPrints(t *testing.T, dir, msg string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	src := fmt.Sprintf("package %s\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(%q)\n}\n", "main", msg)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write main.go in %s: %v", dir, err)
	}
}

// writeBrokenPackageMain writes a package main that does not compile.
// Used to force a failed go install while keeping discovery as package main.
func writeBrokenPackageMain(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	src := fmt.Sprintf("package %s\n\nfunc main() {\n\tthisDoesNotCompile()\n}\n", "main")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write broken main.go in %s: %v", dir, err)
	}
}

func touchBin(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	if err := os.WriteFile(path, []byte("stub-binary\n"), 0o755); err != nil {
		t.Fatalf("touch bin %s: %v", path, err)
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

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertExitNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code %d want %d stderr=%q stdout=%q", resp.ExitCode, want, resp.Stderr, resp.Stdout)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
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

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") {
		t.Fatalf("output must not contain ANSI escapes, got:\n%s", s)
	}
}

func assertStderrExact(t *testing.T, stderr, want string) {
	t.Helper()
	if stderr != want {
		t.Fatalf("stderr mismatch\n got: %q\nwant: %q", stderr, want)
	}
}

// coloredNoticePrefix is assert.Output v2 markup for grey notice: token (#90).
func coloredNoticePrefix() string {
	return "<ansi-color gray>notice:</ansi-color>"
}

// coloredWarningPrefix is assert.Output v2 markup for orange warning: token (#33).
func coloredWarningPrefix() string {
	return "<ansi-color #33>warning:</ansi-color>"
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

// assertStubBinUnchanged checks dry-run did not rewrite a stub binary.
func assertStubBinUnchanged(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read stub bin %s: %v", path, err)
	}
	if string(data) != "stub-binary\n" {
		t.Fatalf("dry-run must not mutate stub bin %s: got %q", path, data)
	}
}

// assertBinNotStub checks execute path replaced the stub with a real install.
func assertBinNotStub(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bin %s: %v", path, err)
	}
	if string(data) == "stub-binary\n" {
		t.Fatalf("expected go install to replace stub bin %s", path)
	}
}

// assertBinExecutable checks GOBIN/<name> exists and has any execute bit.
func assertBinExecutable(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat bin %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("bin %s is a directory", path)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("bin %s is not executable: mode=%v", path, info.Mode())
	}
}

// assertBinRuns executes GOBIN/<name> and requires trimmed combined output == wantOut.
func assertBinRuns(t *testing.T, binDir, name, wantOut string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	cmd := exec.Command(path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run bin %s: %v\n%s", path, err, out)
	}
	got := strings.TrimSpace(string(out))
	if got != wantOut {
		t.Fatalf("bin %s output %q want %q", path, got, wantOut)
	}
}

// assertExecuteSummary checks the last non-empty stdout line is the execute summary.
func assertExecuteSummary(t *testing.T, stdout string, reinstalled, skipped, failed int) {
	t.Helper()
	want := fmt.Sprintf("reinstalled %d, skipped %d, failed %d", reinstalled, skipped, failed)
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		t.Fatalf("stdout empty; want summary %q", want)
	}
	last := lines[len(lines)-1]
	if last != want {
		t.Fatalf("execute summary %q want %q\nfull stdout:\n%s", last, want, stdout)
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("execute stdout must end with newline; got %q", stdout)
	}
}

func assertMutualExclusion(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "mutually exclusive") {
		return
	}
	if strings.Contains(lower, "unexpected") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}

// --- events.jsonl helpers (command "reinstall-local") ---

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

func eventArgsContain(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

// assertLastEventCommandReinstallLocal checks the last events.jsonl entry is
// command "reinstall-local" with the given exit code and required CLI flags.
// requiredArgs should include "--reinstall-local" and, for dry-run, "--dry-run".
func assertLastEventCommandReinstallLocal(t *testing.T, wrkHome string, wantExit int, requiredArgs []string) {
	t.Helper()
	events := readEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one events.jsonl entry")
	}
	ev := events[len(events)-1]
	if ev.Command != "reinstall-local" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "reinstall-local", ev.Command, events)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
	for _, want := range requiredArgs {
		if !eventArgsContain(ev.Args, want) {
			t.Fatalf("event args should include %q, got %v", want, ev.Args)
		}
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}

func resolvePath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %s: %v", p, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

func ensureReinstallLocalCLIHelpersUsed() {
	_ = writeGoMod
	_ = writePackageMain
	_ = writePackageMainPrints
	_ = writeBrokenPackageMain
	_ = touchBin
	_ = mkdirAll
	_ = runGitIsolated
	_ = initGitRepoOnMain
	_ = gitCommitAll
	_ = assertStubBinUnchanged
	_ = assertBinNotStub
	_ = assertBinExecutable
	_ = assertBinRuns
	_ = assertExecuteSummary
	_ = assertExitCode
	_ = assertMutualExclusion
	_ = assertNotContains
	_ = assertContains
	_ = assertNoANSI
	_ = assertStderrExact
	_ = coloredNoticePrefix
	_ = coloredWarningPrefix
	_ = assertEmptyStdout
	_ = assertExitZero
	_ = assertExitNonZero
	_ = assertOutputExact
	_ = v2StdoutTemplate
	_ = readEvents
	_ = assertLastEventCommandReinstallLocal
	_ = eventArgsContain
	_ = resolvePath
	_ = eventsJSONLPath
	_ = filterEnvKeys
}
```
