# Scenario

**Feature**: wrk --main opens a nested interactive shell in the main repository for the current checkout

```
# resolve main from process cwd (ShowToplevel + ResolveMainRepo)
wrk --main

# already at main root
cwd == mainRepo -> stderr notice; no shell; exit 0

# otherwise (main subdir / linked worktree root / linked worktree subdir)
cwd != mainRepo
  -> shell/interactive.LoginInteractive(mainRepo, Base(mainRepo), "WRK_SHELL=1")
  -> always nested shell (ignore WRK_FOLLOWUP_FILE)
  -> minimal launch UX (no install hint, no stdout path)
  -> wrk exit = shell exit
```

## Preconditions

- Shared wrk root harness builds the session `wrk` binary and isolates `WRK_HOME`.
- Feature hangs off root `Request`/`Response`/`Run` — do **not** redefine them here.
- **Launch leaves must install a fake `bash` on PATH** (see `installFakeBash`) so
  `shell/interactive` / `bash.Login` cannot hang CI. Fake shell records cwd/args to
  `req.FakeShellLog` and exits with `req.FakeShellExit` (default 0).
- Follow-up env is **not** set for normal launch leaves (`UseFollowupEnv` false);
  `followup-ignored/` deliberately opens the channel to prove `--main` still nests a shell.
- Invocation is only `wrk --main` (no path positional, no basename).

## Steps

1. Descendants create a git layout (or non-git cwd for error leaves), set `RepoDir`,
   and set `Args` to include `--main` (plus exclusive flags / extra args for errors).
2. Successful launch leaves call `installFakeBash`.
3. `Run` (root) exports harness env via `wrkEnv` / `appendCDEnv` (fake shell PATH/SHELL).

## Context

- Already-at-root notice is on **stderr** and mentions the main repo path.
- Launch: empty/minimal stdout; no `wrk --bash-integration --install` hint.
- Shell cwd is always the resolved **main repo root** (not linked worktree path).
- Mutual exclusion same family as `--cd`; extra positionals → unexpected arguments.
- Event `command: "main"`; `args` include `--main`.
- Fake-shell protocol (same as `cd/`): `WRK_FAKE_SHELL_LOG`, `WRK_FAKE_SHELL_EXIT`,
  `SHELL`, PATH prefix with executable named `bash`.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type wrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, req *Request) error {
	ensureMainHelpersUsed()
	return nil
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

func initMainRepo(t *testing.T, req *Request) string {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	return mainRepo
}

func initMainRepoSubdir(t *testing.T, req *Request, elems ...string) (mainRepo, sub string) {
	t.Helper()
	mainRepo = initMainRepo(t, req)
	if len(elems) == 0 {
		elems = []string{"pkg", "cmd", "tool"}
	}
	parts := append([]string{mainRepo}, elems...)
	sub = filepath.Join(parts...)
	mkdirAll(t, sub)
	return mainRepo, sub
}

func initLinkedWorktree(t *testing.T, req *Request) (mainRepo, linkedWT string) {
	t.Helper()
	mainRepo = initMainRepo(t, req)
	linkedWT = filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "linked-side", linkedWT)
	req.WtDir = linkedWT
	return mainRepo, linkedWT
}

func initLinkedWorktreeSubdir(t *testing.T, req *Request, elems ...string) (mainRepo, linkedWT, sub string) {
	t.Helper()
	mainRepo, linkedWT = initLinkedWorktree(t, req)
	if len(elems) == 0 {
		elems = []string{"pkg", "nested"}
	}
	parts := append([]string{linkedWT}, elems...)
	sub = filepath.Join(parts...)
	mkdirAll(t, sub)
	return mainRepo, linkedWT, sub
}

// installFakeBash places a non-interactive bash shim first on PATH so
// LoginInteractive cannot hang. Call from every successful launch leaf (and
// already-at-root to detect accidental launch).
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

// enableFollowupChannel sets WRK_FOLLOWUP_FILE so leaves can prove --main ignores it.
func enableFollowupChannel(t *testing.T, req *Request) {
	t.Helper()
	req.FollowupFile = filepath.Join(req.WorkRoot, "followup.txt")
	req.UseFollowupEnv = true
}

func setMainArgs(req *Request, extra ...string) {
	req.Args = append([]string{"--main"}, extra...)
	req.TargetDir = ""
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

func assertFollowupEmpty(t *testing.T, req *Request) {
	t.Helper()
	if req.FollowupFile == "" {
		return
	}
	got := readFollowupFile(t, req.FollowupFile)
	if got != "" {
		t.Fatalf("follow-up should be empty ( --main never writes in-place cd), got %q", got)
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

func assertFakeShellNotLaunched(t *testing.T, req *Request) {
	t.Helper()
	if req.FakeShellLog == "" {
		t.Fatal("FakeShellLog not set (installFakeBash required to detect accidental launch)")
	}
	data, err := os.ReadFile(req.FakeShellLog)
	if err != nil {
		t.Fatalf("read fake shell log: %v", err)
	}
	if strings.Contains(string(data), "cwd=") {
		t.Fatalf("fake shell should not have launched; log=%q", data)
	}
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func assertNoInstallHint(t *testing.T, stderr string) {
	t.Helper()
	if strings.Contains(stderr, "wrk --bash-integration --install") {
		t.Fatalf("minimal launch UX: stderr must not mention bash-integration install hint; got %q", stderr)
	}
}

func assertMinimalLaunchUX(t *testing.T, resp *Response) {
	t.Helper()
	assertEmptyStdout(t, resp.Stdout)
	assertNoInstallHint(t, resp.Stderr)
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

func assertLastEventCommandMain(t *testing.T, wrkHome string, wantExit int) {
	t.Helper()
	events := readEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	ev := events[len(events)-1]
	if ev.Command != "main" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "main", ev.Command, events)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
	foundMain := false
	for _, a := range ev.Args {
		if a == "--main" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Fatalf("event args should include --main, got %v", ev.Args)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}

func ensureMainHelpersUsed() {
	_ = resolvePath
	_ = initMainRepo
	_ = initMainRepoSubdir
	_ = initLinkedWorktree
	_ = initLinkedWorktreeSubdir
	_ = installFakeBash
	_ = enableFollowupChannel
	_ = setMainArgs
	_ = readFollowupFile
	_ = assertFollowupEmpty
	_ = assertFakeShellCwd
	_ = assertFakeShellLaunched
	_ = assertFakeShellNotLaunched
	_ = assertEmptyStdout
	_ = assertNoInstallHint
	_ = assertMinimalLaunchUX
	_ = eventsJSONLPath
	_ = readEvents
	_ = assertLastEventCommandMain
}
```
