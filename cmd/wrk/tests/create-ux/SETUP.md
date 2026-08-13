# Scenario

**Feature**: create-mode window / terminal / agent UX driven by config + flags

```
# native create always first
myrepo -> wrk [UX flags] [-t task]
  -> print worktree path\n
  -> optional space.CreateAndActivate (window)
  -> optional iterm2.OpenConfig (terminal ± agent follow-up)
  -> optional agent-run in current process (agent without terminal)

# effective merge
plain create (no <target-dir>): config create.* + CLI flags; --no-* clears axis
create-with-target-dir: skip config create.*; CLI flags only (silent)
--no-config: skip $WRK_HOME/config.json entirely (CLI flags still apply)
window on implies terminal new (after flag apply); legacy create.interceptor ignored
```

## Preconditions

- Git available; isolated `{WRK_HOME}`; `WRK_DATE=2026-06-30`.
- UX mocks via env (implementer must honor space hook; iterm2 package already supports script-out):
  - `WRK_SPACE_INVOKE_LOG` — when set, wrk must log each `CreateAndActivate` (one line `CreateAndActivate`) and **not** run real Mission Control AX; settle must not sleep in tests.
  - `WRK_SPACE_FAIL=max-desktops` — with the invoke log hook, return `space.ErrMaxDesktops` after logging (hermetic max-Desktop capacity soft-fail).
  - `DOT_PKGS_SPACE_GOOS` — platform override for space package (`darwin` / `linux`).
  - `KOOL_ITERM2_SCRIPT_OUT` — iterm2 writes AppleScript here instead of calling real osascript.
  - `KOOL_ITERM2_INSTALLED=1` — pretend iTerm2 is installed.
  - `KOOL_ITERM2_GOOS=darwin` — platform for iterm2 (set `linux` only when testing terminal platform errors).
- Fake `agent-run` on `PATH` (`Request.PathPrepend`) logs argv (ARGC/LEN framing) and cwd.

## Steps

- Grouping installs repo + mock env; leaves set flags/config and assert path, mock logs, and exit.
- Default happy path forces darwin mocks so CI is hermetic on any host OS.

## Context

- Agent default argv: `agent-run run --dir <abs-worktree> --session-id-from-prompt --no-submit --open --color --agent-runner=grok-tty <prompt>`
  (`--dir` may appear immediately after `run`; space form preferred). Always injects `--color` even if config `create.agent.args` omits it.
  Long prompts (`agentrunapi.PromptFileSpillMinRunes`, 600 runes) and follow-ups that would exceed iTerm write-text SafeMax replace the positional prompt with `--prompt-file=<abs>` (file body is the full prompt).
- Default prompt template: `/brainstorm ${task}` (empty task → empty substitution).
- Terminal+agent: agent only as iTerm follow-up command string; outer wrk must **not** exec agent-run.
- In-process agent: `--dir` is the workspace source of truth; process cwd of agent-run **need not** equal worktree.
- Parent follow-up after create: when agent and/or terminal UX is active, skip home-gated
  auto-cd (`writeFollowupCDIfCwdIsHome`) unless `--force-cd`. Bare create still home-gates.
- Pipeline order: create → window → terminal-or-agent → `--exec` → follow-up cd.
- Reuses root `Request`/`Run`; paths for logs live under `{WorkRoot}`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

const (
	envSpaceInvokeLog   = "WRK_SPACE_INVOKE_LOG"
	envSpaceGOOS        = "DOT_PKGS_SPACE_GOOS"
	envItermScriptOut   = "KOOL_ITERM2_SCRIPT_OUT"
	envItermInstalled   = "KOOL_ITERM2_INSTALLED"
	envItermGOOS        = "KOOL_ITERM2_GOOS"
	envFakeAgentRunLog  = "FAKE_AGENT_RUN_LOG"
	envFakeAgentRunCwd  = "FAKE_AGENT_RUN_CWD"
	fakeAgentRunName    = "agent-run"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureCreateUXHelpersUsed()
	return nil
}

func uxSpaceLogPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "space-invoke.log")
}

func uxItermScriptPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "iterm-script.applescript")
}

func uxAgentRunLogPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "agent-run-argv.log")
}

func uxAgentRunCwdPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "agent-run.cwd")
}

func setupMainRepoForCreateUX(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	cloneMainGoModFromSeed(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	return mainRepo
}

func wantCreateUXWorktree(req *Request) string {
	return worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
}

func wantCreateUXWorktreeWithTask(req *Request, task string) string {
	return worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify(task), 0)
}

// installCreateUXMocks installs hermetic space/iterm/agent-run mocks for this leaf.
// goos defaults to "darwin". Pass "linux" to force platform errors on window/terminal.
func installCreateUXMocks(t *testing.T, req *Request, goos string) {
	t.Helper()
	if goos == "" {
		goos = "darwin"
	}
	spaceLog := uxSpaceLogPath(req)
	itermOut := uxItermScriptPath(req)
	agentLog := uxAgentRunLogPath(req)
	agentCwd := uxAgentRunCwdPath(req)
	// Truncate logs so "not invoked" asserts are reliable.
	writeFile(t, spaceLog, "")
	writeFile(t, itermOut, "")
	writeFile(t, agentLog, "")
	writeFile(t, agentCwd, "")

	req.ExtraEnv = append(req.ExtraEnv,
		envSpaceInvokeLog+"="+spaceLog,
		envSpaceGOOS+"="+goos,
		envItermScriptOut+"="+itermOut,
		envItermInstalled+"=1",
		envItermGOOS+"="+goos,
		envFakeAgentRunLog+"="+agentLog,
		envFakeAgentRunCwd+"="+agentCwd,
	)
	installFakeAgentRun(t, req)
}

func installFakeAgentRun(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "ux-bin")
	mkdirAll(t, binDir)
	req.PathPrepend = binDir
	// ARGC/LEN framing (same as historical interceptor fake); also record cwd.
	body := `#!/bin/sh
log="${FAKE_AGENT_RUN_LOG:-}"
cwdlog="${FAKE_AGENT_RUN_CWD:-}"
if [ -n "$cwdlog" ]; then
  pwd > "$cwdlog"
fi
if [ -n "$log" ]; then
  {
    cmd_name=$(basename "$0")
    printf 'ARGC %s\n' "$(($# + 1))"
    for a in "$cmd_name" "$@"; do
      len=$(printf '%s' "$a" | wc -c | tr -d ' \t')
      printf 'LEN %s\n' "$len"
      printf '%s' "$a"
      printf '\n'
    done
  } > "$log"
fi
printf 'agent-run-ok\n'
exit 0
`
	fake := filepath.Join(binDir, fakeAgentRunName)
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake agent-run: %v", err)
	}
}

func writeCreateUXConfig(t *testing.T, wrkHome string, cfg map[string]interface{}) {
	t.Helper()
	root := map[string]interface{}{
		"version": 1,
		"create":  cfg,
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		t.Fatalf("marshal create ux config: %v", err)
	}
	data = append(data, '\n')
	writeFile(t, filepath.Join(wrkHome, "config.json"), string(data))
}

func writeFullCreateUXConfig(t *testing.T, wrkHome string) {
	t.Helper()
	writeCreateUXConfig(t, wrkHome, map[string]interface{}{
		"window":   map[string]interface{}{"mode": "new"},
		"terminal": map[string]interface{}{"mode": "new"},
		"agent": map[string]interface{}{
			"enabled":          true,
			"runner":           "grok-tty",
			"prompt_template":  "/brainstorm ${task}",
			"args":             []string{"--session-id-from-prompt", "--no-submit", "--open", "--color"},
		},
	})
}

func writeInterceptorOnlyConfig(t *testing.T, wrkHome string) {
	t.Helper()
	writeCreateUXConfig(t, wrkHome, map[string]interface{}{
		"interceptor": map[string]interface{}{
			"enabled": true,
			"argv":    []string{"true"},
		},
	})
}

func readFileEmptyOK(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func assertSpaceNotInvoked(t *testing.T, req *Request) {
	t.Helper()
	got := strings.TrimSpace(readFileEmptyOK(t, uxSpaceLogPath(req)))
	if got != "" {
		t.Fatalf("space should not be invoked; log=%q", got)
	}
}

func assertSpaceInvokedOnce(t *testing.T, req *Request) {
	t.Helper()
	got := strings.TrimSpace(readFileEmptyOK(t, uxSpaceLogPath(req)))
	lines := strings.Split(got, "\n")
	var nonempty []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			nonempty = append(nonempty, strings.TrimSpace(ln))
		}
	}
	if len(nonempty) != 1 || nonempty[0] != "CreateAndActivate" {
		t.Fatalf("space log want single CreateAndActivate line, got %q", got)
	}
}

func assertItermNotInvoked(t *testing.T, req *Request) {
	t.Helper()
	got := strings.TrimSpace(readFileEmptyOK(t, uxItermScriptPath(req)))
	if got != "" {
		t.Fatalf("iterm should not be invoked; script=%q", got)
	}
}

func assertItermInvokedAtPath(t *testing.T, req *Request, wtPath string) string {
	t.Helper()
	script := readFileEmptyOK(t, uxItermScriptPath(req))
	if strings.TrimSpace(script) == "" {
		t.Fatal("expected iterm AppleScript written to KOOL_ITERM2_SCRIPT_OUT")
	}
	// OpenConfig EvalSymlinks the dir; compare both raw and resolved forms.
	needles := []string{wtPath}
	if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
		needles = append(needles, resolved)
	}
	found := false
	for _, n := range needles {
		if strings.Contains(script, n) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("iterm script missing worktree path %q (or symlink form); script:\n%s", wtPath, script)
	}
	return script
}

func assertItermModeForceNew(t *testing.T, script string) {
	t.Helper()
	if !strings.Contains(script, `create window with default profile`) {
		t.Fatalf("ForceNew script must create a window; got:\n%s", script)
	}
	// ForceNew skips session scan — no matchingWindow variable.
	if strings.Contains(script, "matchingWindow") {
		t.Fatalf("ForceNew script must not scan matchingWindow; got:\n%s", script)
	}
}

func assertItermModeSmart(t *testing.T, script string) {
	t.Helper()
	if !strings.Contains(script, "matchingWindow") {
		t.Fatalf("Smart script should scan matchingWindow; got:\n%s", script)
	}
	if !strings.Contains(script, `create tab with default profile`) {
		t.Fatalf("Smart script should support create tab; got:\n%s", script)
	}
}

func assertItermModeReuse(t *testing.T, script string) {
	t.Helper()
	if !strings.Contains(script, "matchingSession") {
		t.Fatalf("Reuse script should reference matchingSession; got:\n%s", script)
	}
}

func readAgentRunArgs(t *testing.T, req *Request) []string {
	t.Helper()
	data := readFileEmptyOK(t, uxAgentRunLogPath(req))
	if strings.TrimSpace(data) == "" {
		return nil
	}
	return parseArgcLenLog(t, data)
}

func parseArgcLenLog(t *testing.T, s string) []string {
	t.Helper()
	if !strings.HasPrefix(s, "ARGC ") {
		t.Fatalf("agent-run log missing ARGC header: %q", s)
	}
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) < 2 {
		t.Fatalf("agent-run log truncated: %q", s)
	}
	var argc int
	if _, err := fmt.Sscanf(lines[0], "ARGC %d", &argc); err != nil {
		t.Fatalf("parse ARGC: %v in %q", err, lines[0])
	}
	rest := []byte(lines[1])
	var args []string
	for i := 0; i < argc; i++ {
		nl := -1
		for j, b := range rest {
			if b == '\n' {
				nl = j
				break
			}
		}
		if nl < 0 {
			t.Fatalf("agent-run log: missing LEN for arg %d", i)
		}
		header := string(rest[:nl])
		var n int
		if _, err := fmt.Sscanf(header, "LEN %d", &n); err != nil {
			t.Fatalf("parse LEN for arg %d: %v in %q", i, err, header)
		}
		rest = rest[nl+1:]
		if len(rest) < n {
			t.Fatalf("agent-run log: arg %d short payload", i)
		}
		args = append(args, string(rest[:n]))
		rest = rest[n:]
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
	}
	return args
}

func assertAgentRunNotInvoked(t *testing.T, req *Request) {
	t.Helper()
	args := readAgentRunArgs(t, req)
	if len(args) != 0 {
		t.Fatalf("outer agent-run should not run; argv=%v", args)
	}
}

func assertAgentRunInvoked(t *testing.T, req *Request, wtPath, task string) []string {
	t.Helper()
	args := readAgentRunArgs(t, req)
	if len(args) == 0 {
		t.Fatal("expected outer agent-run invocation")
	}
	if args[0] != fakeAgentRunName {
		t.Fatalf("argv0: want %q, got %q", fakeAgentRunName, args[0])
	}
	// Shape: agent-run run --dir <wt> <args...> --agent-runner=<runner> <prompt>
	if len(args) < 3 || args[1] != "run" {
		t.Fatalf("argv should start with agent-run run …, got %v", args)
	}
	joined := strings.Join(args, "\x00")
	for _, need := range []string{"--session-id-from-prompt", "--no-submit", "--open", "--color"} {
		if !containsArg(args, need) {
			t.Fatalf("argv missing %q: %v", need, args)
		}
	}
	if !strings.Contains(joined, "grok-tty") {
		t.Fatalf("argv missing grok-tty runner: %v", args)
	}
	assertAgentArgvHasDir(t, args, wtPath)
	wantPrompt := "/brainstorm"
	if task != "" {
		wantPrompt = "/brainstorm " + task
	}
	// prompt is last arg (or last non-flag)
	last := args[len(args)-1]
	if last != wantPrompt && !strings.HasSuffix(last, wantPrompt) {
		// allow template empty-task edge: "/brainstorm " with trailing space stripped
		if !(task == "" && (last == "/brainstorm" || last == "/brainstorm ")) {
			t.Fatalf("prompt token: want %q, got last=%q full=%v", wantPrompt, last, args)
		}
	}
	// Process cwd of agent-run may differ from worktree; --dir is the workspace source of truth.
	// Still require the fake to have been invoked (cwd log non-empty when mock records it).
	cwd := strings.TrimSpace(readFileEmptyOK(t, uxAgentRunCwdPath(req)))
	if cwd == "" {
		t.Fatal("agent-run cwd log empty (fake agent-run should record pwd)")
	}
	return args
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want || strings.HasPrefix(a, want+"=") {
			return true
		}
	}
	return false
}

// agentArgvDir returns the --dir value from argv (space form or --dir=PATH).
func agentArgvDir(args []string) (string, bool) {
	for i, a := range args {
		if a == "--dir" {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", false
		}
		if strings.HasPrefix(a, "--dir=") {
			return strings.TrimPrefix(a, "--dir="), true
		}
	}
	return "", false
}

func uxPathsEqual(a, b string) bool {
	ca, cb := a, b
	if r, err := filepath.EvalSymlinks(a); err == nil {
		ca = r
	}
	if r, err := filepath.EvalSymlinks(b); err == nil {
		cb = r
	}
	return filepath.Clean(ca) == filepath.Clean(cb)
}

func assertAgentArgvHasDir(t *testing.T, args []string, wtPath string) {
	t.Helper()
	dir, ok := agentArgvDir(args)
	if !ok || strings.TrimSpace(dir) == "" {
		t.Fatalf("agent-run argv missing --dir <worktree>; argv=%v", args)
	}
	if !uxPathsEqual(dir, wtPath) {
		t.Fatalf("agent-run --dir: want worktree %q, got %q (argv=%v)", wtPath, dir, args)
	}
}

func itermFollowUpPromptFilePath(script string) (string, bool) {
	const flag = "--prompt-file="
	for _, line := range strings.Split(script, "\n") {
		if !strings.Contains(line, "write text") {
			continue
		}
		idx := strings.Index(line, flag)
		if idx < 0 {
			continue
		}
		rest := line[idx+len(flag):]
		end := len(rest)
		for i, r := range rest {
			if r == '"' || r == '\'' || r == ' ' {
				end = i
				break
			}
		}
		path := rest[:end]
		if path != "" {
			return path, true
		}
	}
	return "", false
}

func assertItermFollowUpUsesPromptFile(t *testing.T, script, wantPrompt string) string {
	t.Helper()
	path, ok := itermFollowUpPromptFilePath(script)
	if !ok {
		t.Fatalf("iterm follow-up should use --prompt-file; script:\n%s", script)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("--prompt-file path must be absolute; got %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spill %q: %v", path, err)
	}
	got := strings.TrimSpace(string(data))
	want := strings.TrimSpace(wantPrompt)
	if got != want {
		t.Fatalf("spill body: want %q got %q (path=%s)", want, got, path)
	}
	if strings.Contains(script, want) {
		t.Fatalf("iterm script must not embed long prompt body; script:\n%s", script)
	}
	return path
}

func agentArgvPromptFile(args []string) (string, bool) {
	for i, a := range args {
		if a == "--prompt-file" {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", false
		}
		if strings.HasPrefix(a, "--prompt-file=") {
			return strings.TrimPrefix(a, "--prompt-file="), true
		}
	}
	return "", false
}

func assertAgentRunInvokedWithPromptFile(t *testing.T, req *Request, wtPath, wantPrompt string) []string {
	t.Helper()
	args := readAgentRunArgs(t, req)
	if len(args) == 0 {
		t.Fatal("expected outer agent-run invocation")
	}
	assertAgentArgvHasDir(t, args, wtPath)
	path, ok := agentArgvPromptFile(args)
	if !ok {
		t.Fatalf("agent-run argv missing --prompt-file; argv=%v", args)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("--prompt-file path must be absolute; got %q argv=%v", path, args)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spill %q: %v", path, err)
	}
	got := strings.TrimSpace(string(data))
	want := strings.TrimSpace(wantPrompt)
	if got != want {
		t.Fatalf("spill body: want %q got %q (path=%s argv=%v)", want, got, path, args)
	}
	last := args[len(args)-1]
	if last == want {
		t.Fatalf("positional prompt must be omitted when using --prompt-file; argv=%v", args)
	}
	return args
}

func assertItermFollowUpHasAgentRun(t *testing.T, script, wtPath, task string) {
	t.Helper()
	if !strings.Contains(script, "agent-run") {
		t.Fatalf("iterm follow-up should contain agent-run; script:\n%s", script)
	}
	if !strings.Contains(script, "write text") {
		t.Fatalf("iterm script should write text follow-ups; script:\n%s", script)
	}
	if !strings.Contains(script, "grok-tty") {
		t.Fatalf("iterm follow-up should include grok-tty; script:\n%s", script)
	}
	if !strings.Contains(script, "--color") {
		t.Fatalf("iterm follow-up should include --color; script:\n%s", script)
	}
	if !strings.Contains(script, "--dir") {
		t.Fatalf("iterm follow-up should include --dir; script:\n%s", script)
	}
	// Worktree path (raw or shell-quoted) must appear next to --dir usage.
	needles := []string{wtPath}
	if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
		needles = append(needles, resolved)
	}
	foundWT := false
	for _, n := range needles {
		if n != "" && (strings.Contains(script, n) || strings.Contains(script, shellSafeQuoteUX(n))) {
			foundWT = true
			break
		}
	}
	if !foundWT {
		t.Fatalf("iterm follow-up should include worktree path for --dir %q; script:\n%s", wtPath, script)
	}
	if task != "" && !strings.Contains(script, task) && !strings.Contains(script, shellSafeQuoteUX(task)) {
		// task may be shell-quoted inside the command line
		t.Fatalf("iterm follow-up should carry task %q (raw or shell-safe); script:\n%s", task, script)
	}
}

func shellSafeQuoteUX(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func readFollowupFileUX(t *testing.T, req *Request) string {
	t.Helper()
	if req.FollowupFile == "" {
		t.Fatal("FollowupFile empty")
	}
	return readFileEmptyOK(t, req.FollowupFile)
}

func assertFollowupEmptyUX(t *testing.T, req *Request) {
	t.Helper()
	if req.FollowupFile == "" {
		return
	}
	got := readFollowupFileUX(t, req)
	if strings.TrimSpace(got) != "" {
		t.Fatalf("follow-up should be empty, got %q", got)
	}
}

func assertFollowupCDUX(t *testing.T, req *Request, wantAbs string) {
	t.Helper()
	got := readFollowupFileUX(t, req)
	candidates := []string{filepath.Clean(wantAbs)}
	if r, err := filepath.EvalSymlinks(wantAbs); err == nil {
		candidates = append(candidates, r)
	}
	trimmed := strings.TrimSpace(got)
	matched := false
	for _, c := range candidates {
		if trimmed == "cd "+c {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("follow-up content: want cd <worktree>, got %q (candidates %v)", got, candidates)
	}
	if got != "" && !strings.HasSuffix(got, "\n") {
		t.Fatalf("follow-up file should end with newline; got %q", got)
	}
}

// setupCreateUXFromFakeHome prepares create from FakeHome with WRK_FOLLOWUP_FILE open
// so home-gated parent auto-cd can be observed. Main repo is passed as TargetDir.
func setupCreateUXFromFakeHome(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	req.FakeHome = filepath.Join(req.WorkRoot, "home")
	mkdirAll(t, req.FakeHome)
	req.RepoDir = req.FakeHome
	req.TargetDir = mainRepo
	req.FollowupFile = filepath.Join(req.WorkRoot, "followup.txt")
	req.UseFollowupEnv = true
	return mainRepo
}

func assertNativeCreateOK(t *testing.T, req *Request, resp *Response, err error, wtPath string) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, v2StdoutTemplate(wtPath))
	assertFileExists(t, wtPath)
	assertGitFileIsWorktreeLink(t, wtPath)
}

// assert package import used by assert.Output — referenced via ensure.
func ensureCreateUXHelpersUsed() {
	_ = uxSpaceLogPath
	_ = uxItermScriptPath
	_ = uxAgentRunLogPath
	_ = uxAgentRunCwdPath
	_ = setupMainRepoForCreateUX
	_ = wantCreateUXWorktree
	_ = wantCreateUXWorktreeWithTask
	_ = installCreateUXMocks
	_ = installFakeAgentRun
	_ = writeCreateUXConfig
	_ = writeFullCreateUXConfig
	_ = writeInterceptorOnlyConfig
	_ = readFileEmptyOK
	_ = assertSpaceNotInvoked
	_ = assertSpaceInvokedOnce
	_ = assertItermNotInvoked
	_ = assertItermInvokedAtPath
	_ = assertItermModeForceNew
	_ = assertItermModeSmart
	_ = assertItermModeReuse
	_ = readAgentRunArgs
	_ = parseArgcLenLog
	_ = assertAgentRunNotInvoked
	_ = assertAgentRunInvoked
	_ = containsArg
	_ = agentArgvDir
	_ = uxPathsEqual
	_ = assertAgentArgvHasDir
	_ = assertItermFollowUpHasAgentRun
	_ = itermFollowUpPromptFilePath
	_ = assertItermFollowUpUsesPromptFile
	_ = agentArgvPromptFile
	_ = assertAgentRunInvokedWithPromptFile
	_ = shellSafeQuoteUX
	_ = assertNativeCreateOK
	_ = envSpaceInvokeLog
	_ = envFakeAgentRunLog
	_ = readFollowupFileUX
	_ = assertFollowupEmptyUX
	_ = assertFollowupCDUX
	_ = setupCreateUXFromFakeHome
}
```
