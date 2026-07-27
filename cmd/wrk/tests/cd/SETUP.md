# Scenario

**Feature**: wrk --cd jump into a directory (in-place follow-up or fallback interactive shell)

```
# path resolve (basename via projects.json; abs/relative via Abs+stat)
wrk --cd <path|basename>  OR  wrk <path|basename> --cd

# Branch A — channel open (bash integration runtime)
WRK_FOLLOWUP_FILE set -> writeFollowupCD(false, abs); empty stdout; no shell

# Branch B — channel closed
WRK_FOLLOWUP_FILE unset -> stderr install hint; stdout abs\n; shell/interactive.LoginInteractive
```

## Preconditions

- Shared wrk root harness builds the session `wrk` binary and isolates `WRK_HOME`.
- **Fallback leaves must install a fake `bash` on PATH** (see `installFakeBash`) so
  `shell/interactive` / `bash.Login` cannot hang CI. Fake shell records cwd/args to
  `req.FakeShellLog` and exits with `req.FakeShellExit` (default 0).
- In-place leaves set `UseFollowupEnv` + `FollowupFile`; they must **not** launch a shell.
- Path is a positional with `Bool("--cd")` (not a String flag): `wrk --cd PATH` and
  `wrk PATH --cd` are both valid.

## Steps

1. Descendants set target path, CLI form (`Args` / `TargetDir`), follow-up env, and
   optional fake shell.
2. `Run` (root) prepares empty `FollowupFile` and exports harness env via `wrkEnv`.

## Context

- Follow-up line format: exactly `cd /absolute/path\n` (expanded abs, never basename).
- Fallback stderr must mention `wrk --bash-integration --install`.
- Fallback stdout: one absolute path line + trailing `\n`.
- In-place stdout: empty (no bytes).
- Mutual exclusion like `--where`; also exclusive with `--no-cd`.
- Event `command: "cd"`; `args` include `--cd`.
- Fake-shell protocol: `WRK_FAKE_SHELL_LOG`, `WRK_FAKE_SHELL_EXIT`, `SHELL`, PATH prefix
  with a executable named `bash` (detect.Shell basename → bash → exec.LookPath("bash")).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"github.com/xhd2015/doctest/session"
)

const cdBasename = "myrepo"

type wrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureCDHelpersUsed()
	return nil
}

func cdAbsTarget(t *testing.T, req *Request, name string) string {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, name)
	mkdirAll(t, dir)
	return resolvePath(t, dir)
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	mkdirAll(t, cwd)
	return cwd
}

func initSavedGitRepo(t *testing.T, workRoot, parent, basename string) string {
	t.Helper()
	repoPath := filepath.Join(workRoot, parent, basename)
	initGitRepoOnMain(t, repoPath)
	return repoPath
}

func recordSavedProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", repoPath)
}

func resolvePath(t *testing.T, path string) string {
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

func sortedSavedPaths(t *testing.T, paths ...string) []string {
	t.Helper()
	var out []string
	for _, p := range paths {
		out = append(out, resolvePath(t, p))
	}
	sort.Strings(out)
	return out
}

// enableInPlaceChannel sets WRK_FOLLOWUP_FILE harness for Branch A (no interactive shell).
func enableInPlaceChannel(t *testing.T, req *Request) {
	t.Helper()
	req.FollowupFile = filepath.Join(req.WorkRoot, "followup.txt")
	req.UseFollowupEnv = true
}

// installFakeBash places a non-interactive bash shim first on PATH so fallback
// LoginInteractive cannot hang. Call from every successful fallback leaf.
func installFakeBash(t *testing.T, req *Request, exitCode int) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "fake-shell-bin")
	mkdirAll(t, binDir)
	logPath := filepath.Join(req.WorkRoot, "fake-shell-log.txt")
	// Truncate prior log.
	writeFile(t, logPath, "")
	fake := filepath.Join(binDir, "bash")
	// Record cwd + args; honor WRK_FAKE_SHELL_EXIT (harness also sets it).
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

// setCDFlagThenPath: wrk --cd <path>
func setCDFlagThenPath(req *Request, path string) {
	req.Args = []string{"--cd", path}
	req.TargetDir = ""
}

// setCDPathThenFlag: wrk <path> --cd
func setCDPathThenFlag(req *Request, path string) {
	req.TargetDir = path
	req.Args = []string{"--cd"}
}

func readFollowupFile(t *testing.T, path string) string {
	t.Helper()
	if path == "" {
		t.Fatal("FollowupFile empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read follow-up %s: %v", path, err)
	}
	return string(data)
}

func assertFollowupCDLine(t *testing.T, req *Request, wantAbs string) {
	t.Helper()
	wantAbs = resolvePath(t, wantAbs)
	got := readFollowupFile(t, req.FollowupFile)
	want := "cd " + wantAbs + "\n"
	if got != want {
		t.Fatalf("follow-up content:\n want %q\n  got %q", want, got)
	}
}

func assertFollowupEmpty(t *testing.T, req *Request) {
	t.Helper()
	if req.FollowupFile == "" {
		return
	}
	got := readFollowupFile(t, req.FollowupFile)
	if got != "" {
		t.Fatalf("follow-up should be empty, got %q", got)
	}
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

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func assertInstallHint(t *testing.T, stderr string) {
	t.Helper()
	assertContains(t, stderr, "wrk --bash-integration --install")
}

func eventsJSONLPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
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

func assertLastEventCommand(t *testing.T, wrkHome, wantCommand string, wantExit int, wantArgs []string) {
	t.Helper()
	events := readEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	ev := events[len(events)-1]
	if ev.Command != wantCommand {
		t.Fatalf("event command: want %q, got %q", wantCommand, ev.Command)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
	if wantArgs == nil {
		wantArgs = []string{}
	}
	if ev.Args == nil {
		ev.Args = []string{}
	}
	if fmt.Sprint(ev.Args) != fmt.Sprint(wantArgs) {
		t.Fatalf("event args: want %v, got %v", wantArgs, ev.Args)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}

func ensureCDHelpersUsed() {
	_ = cdAbsTarget
	_ = initNeutralCwd
	_ = initSavedGitRepo
	_ = recordSavedProject
	_ = resolvePath
	_ = sortedSavedPaths
	_ = enableInPlaceChannel
	_ = installFakeBash
	_ = setCDFlagThenPath
	_ = setCDPathThenFlag
	_ = readFollowupFile
	_ = assertFollowupCDLine
	_ = assertFollowupEmpty
	_ = assertFakeShellCwd
	_ = assertFakeShellLaunched
	_ = assertEmptyStdout
	_ = assertInstallHint
	_ = eventsJSONLPath
	_ = readEvents
	_ = assertLastEventCommand
}
```
