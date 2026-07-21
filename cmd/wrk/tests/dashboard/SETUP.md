# Scenario

**Feature**: dashboard entry + `--new` create + non-TTY snapshot (P2) + interactive RUN/CANCEL (P3)

```
# bare no-args is dashboard (not create)
myrepo | linked-wt -> wrk
  -> no new worktree under WRK_HOME/worktrees
  -> P2 non-TTY (no action env): static dashboard snapshot; exit 0
  -> P3 TTY or WRK_DASHBOARD_ACTION=*: interactive / forced action path
  -> event command "dashboard" (cancel and non-compose exits)

# --new is the create entry (former bare create)
myrepo -> wrk --new
  -> exit 0; stdout abs worktree path\n
  -> worktree under {WRK_HOME}/worktrees/

# create still available without --new when args select create
myrepo -> wrk <dir> | wrk -t task
  -> create (unchanged)

# --new is exclusive with other modes
wrk --new --done | --list | --status
  -> non-zero; mutually exclusive

# P3 RUN (hermetic): WRK_DASHBOARD_ACTION=run-done|run-merge-back
#   + optional WRK_DASHBOARD_DRY_RUN=1 → real multi-stage compose argv
```

## Preconditions

- Git available for create / bare-no-create / snapshot / interactive leaves that need a repo cwd.
- Isolated `{WRK_HOME}` at `{WorkRoot}/.wrk`; `WRK_DATE=2026-06-30`.
- Reuses root `Request` / `Run` from `cmd/wrk/tests/DOCTEST.md`.

## Steps

- Grouping installs shared helpers; leaves set `req.Args` / `TargetDir` / `TaskDesc` / cwd / `ExtraEnv`.
- P1: create-via-new / mutex / create-without-new / bare-no-create routing.
- P2: `snapshot/` non-TTY fine-grained dashboard View (static; no Bubble Tea keys).
- P3: `interactive/` forced actions via env (no real PTY).
- P3b: `tty/` real PTY via **tty-watch** (Bubble Tea stays alive; keys CANCEL/RUN).
- P4: `docs/` help + skill coverage for dashboard / `--new`.
- P5: `package-tui/` package-boundary leaves for `github.com/xhd2015/wrk/wrkcli/tui`
  (importable / RunDashboard+types exported / no parent import cycle). Classic TDD: RED until implementer.

## Context

- Product: mode name **dashboard** for bare no-args; create entry is **`wrk --new`**.
- Event: bare no-args / cancel → `command: "dashboard"`; `wrk --new` → `command: "create"`; successful RUN compose may record primary mode (`done` / `merge-back`) per existing compose event rules — cancel must stay `dashboard`.
- Snapshot glyphs: **`[•]`** on / **`[ ]`** off / **`[-]`** disabled — **never** `[x]`/`[X]`.
- Stages (fine-grained): Pre **`add changes`** (lowercase label); **gen-commit-msg** with nested **agent-runner** (default **commandcode**); **commit**; Main **MERGE BACK** then **DONE** (MERGE BACK default-selected when linked); After **sync**, **tag-next**, **push**, **reinstall-local**.
- Compact non-TTY snapshot: **no** create-hint (`hint: create a worktree…` / `wrk --new` create tip); Batch line includes **`would run:`**.
- Mutex wording mirrors other modes (`mutually exclusive`).

### P3 hermetic env contract (implementer MUST honor)

| Env | Values / meaning |
|-----|------------------|
| `WRK_DASHBOARD_ACTION` | When set, force the interactive action path **even if stdin is non-TTY** (doctest hook; no PTY). Values: `cancel` \| `run-done` \| `run-merge-back`. Unknown → non-zero clear error. |
| `WRK_DASHBOARD_DRY_RUN` | `1` → inject `--dry-run` into composed RUN argv (plan-only, zero mutations). |
| `WRK_DASHBOARD_COMPOSE_ARGV_LOG` | Path: before executing RUN compose, write composed argv tokens (one token per line preferred; space-joined single line also OK) so tests can assert recipe without parsing all of stdout. |
| `WRK_DASHBOARD_TOGGLES` | Optional comma list `stage=on\|off` (ids: `add-changes`, `gen-commit-msg`, `commit`, `done`, `merge-back`, `sync`, `tag-next`, `push`, `reinstall-local`). **Disabled gates cannot be forced on** (ignore force-on or error; recipe must not include gated flags). |

**Default RUN recipe (DONE, defaults on):**  
`wrk --gen-commit-msg [--add-all if Add changes on] --commit --agent-runner=commandcode --done --sync --tag-next --push --reinstall-local`  
plus `--dry-run` when `WRK_DASHBOARD_DRY_RUN=1`.  
**MERGE BACK:** same with `--merge-back` instead of `--done`.

**CANCEL:** exit 0; no compose mutations; event `command: "dashboard"`.

**Non-TTY without `WRK_DASHBOARD_ACTION`:** P2 static snapshot (unchanged).

### P3b real TTY via tty-watch (Bubble Tea)

Root `Run` uses pipes (not a TTY). Drive the live dashboard with **`tty-watch`**:

```sh
# TERM=dumb: Bubble Tea package init queries OSC background colors; non-responding
# PTYs hang without TERM=dumb (real Terminal.app/iTerm are fine with default TERM).
export TERM=dumb
export TTY_WATCH_HOME=$WorkRoot/tty-watch-home
export WRK_HOME=... WRK_DATE=2026-06-30
tty-watch run --detach --session-id <id> -- env TERM=dumb WRK_HOME=... WRK_DATE=... <wrkbin>
tty-watch snapshot <id>   # must show dashboard content while process still alive
tty-watch send <id> $'q'  # CANCEL
tty-watch snapshot <id>   # eventually [Terminal exited]
tty-watch kill <id>
```

**Keys (TTY Bubble Tea):** row-wise focus only (no per-row Run chip focus). `↑/↓` or `j/k` move; `space` toggles stage include (not disabled); **Enter on stage** = single-phase run; on Main `enter`/`space` selects **MERGE BACK** vs **DONE** (default focus prefers MERGE BACK when linked); on **RUN ALL** `enter` runs batch compose; on CANCEL `enter` cancels; `q`/`Esc` cancel. After stage run / RUN ALL, TUI may re-open (clearInline hard to assert in hermetic).

**Env on TTY RUN path (same as hermetic):** `WRK_DASHBOARD_DRY_RUN=1`, `WRK_DASHBOARD_COMPOSE_ARGV_LOG=<path>` honored after tea quits with RUN.

```go
import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	ensureDashboardHelpersUsed()
	return nil
}

type dashEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func dashboardEventsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readDashboardEvents(t *testing.T, wrkHome string) []dashEvent {
	t.Helper()
	data, err := os.ReadFile(dashboardEventsPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []dashEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev dashEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func setupDashboardMainRepo(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	return mainRepo
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

func wantDashboardCreateWorktree(req *Request) string {
	return worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
}

func wantDashboardCreateWorktreeWithTask(req *Request, task string) string {
	return worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify(task), 0)
}

func worktreesRoot(req *Request) string {
	return filepath.Join(req.WrkHome, "worktrees")
}

// listWorktreeEntries returns basenames under WRK_HOME/worktrees (empty if missing).
func listWorktreeEntries(t *testing.T, req *Request) []string {
	t.Helper()
	root := worktreesRoot(req)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("readdir worktrees: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func assertNoWorktreesCreated(t *testing.T, req *Request) {
	t.Helper()
	names := listWorktreeEntries(t, req)
	if len(names) != 0 {
		t.Fatalf("bare no-args must not create worktrees; found under %s: %v", worktreesRoot(req), names)
	}
	// Expected create path must not exist either (in case layout differs).
	want := wantDashboardCreateWorktree(req)
	if _, err := os.Stat(want); err == nil {
		t.Fatalf("bare no-args must not create worktree path %s", want)
	}
}

func assertCreateOKPath(t *testing.T, req *Request, resp *Response, err error, wtPath string) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, wtPath)
	assertFileExists(t, wtPath)
	assertGitFileIsWorktreeLink(t, wtPath)
}

func assertMutexNewMode(t *testing.T, resp *Response, err error, otherFlag string) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --new + %s, stdout=%q stderr=%q", otherFlag, resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty on mutex error, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "mutually exclusive") && !strings.Contains(se, "exclusive") {
		t.Fatalf("stderr should mention mutual exclusion, got %q", resp.Stderr)
	}
	// Prefer naming the flags involved when the message is specific.
	if !strings.Contains(se, "--new") && !strings.Contains(se, "new") {
		// soft preference: still accept generic exclusive wording
	}
}

// setupDashboardLinkedWorktree creates main + linked worktree (not under WRK_HOME/worktrees).
// Returns linked path; sets MainRepo / WtDir / RepoDir=linked.
func setupDashboardLinkedWorktree(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := setupDashboardMainRepo(t, req)
	linked := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "feature-dash", linked)
	req.WtDir = linked
	req.WtBranch = "feature-dash"
	req.RepoDir = linked
	return linked
}

func markDashboardDirtyUntracked(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "dash-dirty.txt"), "untracked for dashboard\n")
}

// lineContaining returns first stdout line that contains needle (case-sensitive), or "".
func lineContaining(stdout, needle string) string {
	for _, ln := range strings.Split(stdout, "\n") {
		if strings.Contains(ln, needle) {
			return ln
		}
	}
	return ""
}

func stdoutHasAny(stdout string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(stdout, n) {
			return true
		}
	}
	return false
}

func stdoutHasAllFold(stdout string, needles ...string) []string {
	lower := strings.ToLower(stdout)
	var missing []string
	for _, n := range needles {
		if !strings.Contains(lower, strings.ToLower(n)) {
			missing = append(missing, n)
		}
	}
	return missing
}

// assertDashboardSnapshotCore checks P2 non-TTY static dashboard View shape.
// Requires identity, stages, glyphs; MVP: lowercase "add changes", MERGE BACK before DONE,
// Batch "would run", and **no** create-hint / wrk --new tip in snapshot.
func assertDashboardSnapshotCore(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("dashboard snapshot exit: want 0, got %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if strings.TrimSpace(out) == "" {
		t.Fatal("dashboard snapshot stdout empty")
	}
	// Identity
	if !strings.Contains(strings.ToLower(out), "dashboard") {
		t.Fatalf("stdout must identify dashboard mode, got %q", out)
	}
	// Glyphs: allowed set present; forbidden [x]/[X]
	if strings.Contains(out, "[x]") || strings.Contains(out, "[X]") {
		t.Fatalf("dashboard must not use [x]/[X] glyphs; stdout:\n%s", out)
	}
	if !stdoutHasAny(out, "[•]", "[ ]", "[-]") {
		t.Fatalf("stdout must use fine-grained glyphs [•]/[ ]/[-]; got:\n%s", out)
	}
	// Pre / Main / After section labels (case-insensitive)
	if missing := stdoutHasAllFold(out, "Pre", "Main", "After"); len(missing) > 0 {
		t.Fatalf("stdout missing section label(s) %v; got:\n%s", missing, out)
	}
	// Label is lowercase "add changes" (not title-case "Add changes")
	if strings.Contains(out, "Add changes") {
		t.Fatalf("stage label must be lowercase %q, not title-case %q; got:\n%s",
			"add changes", "Add changes", out)
	}
	if !strings.Contains(out, "add changes") {
		t.Fatalf("stdout missing lowercase stage label %q; got:\n%s", "add changes", out)
	}
	// Fine-grained stage rows
	stageNeedles := []string{
		"add changes",
		"gen-commit-msg",
		"agent-runner",
		"commandcode",
		"commit",
		"DONE",
		"MERGE BACK",
		"sync",
		"tag-next",
		"push",
		"reinstall-local",
	}
	// MERGE BACK may appear as MERGE BACK or MERGE-BACK
	var missingStages []string
	lower := strings.ToLower(out)
	for _, n := range stageNeedles {
		nl := strings.ToLower(n)
		if n == "MERGE BACK" {
			if !strings.Contains(lower, "merge back") && !strings.Contains(lower, "merge-back") {
				missingStages = append(missingStages, n)
			}
			continue
		}
		if !strings.Contains(lower, nl) {
			missingStages = append(missingStages, n)
		}
	}
	if len(missingStages) > 0 {
		t.Fatalf("stdout missing stage/label(s) %v; got:\n%s", missingStages, out)
	}
	// Main order: MERGE BACK first, then DONE
	mbIdx := strings.Index(lower, "merge back")
	if mbIdx < 0 {
		mbIdx = strings.Index(lower, "merge-back")
	}
	doneIdx := -1
	off := 0
	for _, line := range strings.Split(out, "\n") {
		ll := strings.ToLower(line)
		if !strings.Contains(ll, "merge") &&
			strings.Contains(ll, "done") &&
			(strings.Contains(line, "[•]") || strings.Contains(line, "[ ]") || strings.Contains(line, "[-]")) {
			doneIdx = off + strings.Index(ll, "done")
			break
		}
		off += len(line) + 1
	}
	if mbIdx >= 0 && doneIdx >= 0 && mbIdx > doneIdx {
		t.Fatalf("Main order: MERGE BACK must appear before DONE; got:\n%s", out)
	}
	// Nested agent-runner under gen-commit-msg: both appear (order soft-checked)
	gcmIdx := strings.Index(lower, "gen-commit-msg")
	arIdx := strings.Index(lower, "agent-runner")
	if gcmIdx >= 0 && arIdx >= 0 && arIdx < gcmIdx {
		// Soft: agent-runner presence already checked.
	}
	// Batch preview line (compact single-frame)
	if !strings.Contains(lower, "would run") {
		t.Fatalf("snapshot Batch line must include would-run preview; got:\n%s", out)
	}
	// No create-hint (MVP: snapshot is not create-mode help)
	if strings.Contains(lower, "create a worktree") ||
		strings.Contains(lower, "hint: create") ||
		strings.Contains(out, "hint: create a worktree with wrk --new") {
		t.Fatalf("snapshot must not show create-hint; got:\n%s", out)
	}
	// No create side effects under WRK_HOME/worktrees from bare dashboard
	assertNoWorktreesCreated(t, req)
	_ = lineContaining
}

// assertAddChangesGlyph finds the add changes row and checks its glyph class.
// wantDisabled: true → row must show [-]; false → row must not be disabled-only (expect [•] or [ ]).
func assertAddChangesGlyph(t *testing.T, stdout string, wantDisabled bool) {
	t.Helper()
	ln := lineContaining(stdout, "add changes")
	if ln == "" {
		for _, line := range strings.Split(stdout, "\n") {
			if strings.Contains(strings.ToLower(line), "add changes") {
				ln = line
				break
			}
		}
	}
	if ln == "" {
		t.Fatalf("no add changes row in stdout:\n%s", stdout)
	}
	hasOn := strings.Contains(ln, "[•]")
	hasOff := strings.Contains(ln, "[ ]")
	hasDis := strings.Contains(ln, "[-]")
	if wantDisabled {
		if !hasDis {
			t.Fatalf("clean tree: add changes should be disabled [-]; line=%q", ln)
		}
		return
	}
	// Dirty: enabled glyph [•] or [ ] (not only disabled)
	if hasDis && !hasOn && !hasOff {
		t.Fatalf("dirty tree: add changes should be enabled ([•] or [ ]), not only [-]; line=%q", ln)
	}
	if !hasOn && !hasOff && !hasDis {
		t.Fatalf("add changes row missing glyph; line=%q", ln)
	}
	if !hasOn && !hasOff {
		t.Fatalf("dirty tree: expected [•] or [ ] on add changes; line=%q", ln)
	}
}

// assertMergeBackDefaultSelected: linked worktree Main — MERGE BACK [•], DONE [ ] (not main-disabled).
func assertMergeBackDefaultSelected(t *testing.T, stdout string) {
	t.Helper()
	mergeLn := ""
	doneLn := ""
	for _, line := range strings.Split(stdout, "\n") {
		ll := strings.ToLower(line)
		if strings.Contains(ll, "merge back") || strings.Contains(ll, "merge-back") {
			if mergeLn == "" {
				mergeLn = line
			}
			continue
		}
		if strings.Contains(ll, "done") && (strings.Contains(line, "[•]") || strings.Contains(line, "[ ]") || strings.Contains(line, "[-]")) {
			if doneLn == "" {
				doneLn = line
			}
		}
	}
	if mergeLn == "" {
		t.Fatalf("no MERGE BACK row in stdout:\n%s", stdout)
	}
	if doneLn == "" {
		t.Fatalf("no DONE row in stdout:\n%s", stdout)
	}
	if !strings.Contains(mergeLn, "[•]") {
		t.Fatalf("linked default: MERGE BACK should be selected [•]; line=%q", mergeLn)
	}
	if strings.Contains(doneLn, "[•]") {
		t.Fatalf("linked default: DONE should not be selected [•]; line=%q", doneLn)
	}
	if !strings.Contains(doneLn, "[ ]") && !strings.Contains(doneLn, "[-]") {
		t.Fatalf("linked default: DONE should show [ ] (or [-]); line=%q", doneLn)
	}
}

// --- P3 interactive / hermetic action hooks ---

const (
	envDashboardAction        = "WRK_DASHBOARD_ACTION"
	envDashboardDryRun        = "WRK_DASHBOARD_DRY_RUN"
	envDashboardComposeArgvLog = "WRK_DASHBOARD_COMPOSE_ARGV_LOG"
	envDashboardToggles       = "WRK_DASHBOARD_TOGGLES"
)

func dashComposeArgvLogPath(req *Request) string {
	return filepath.Join(req.WorkRoot, "dashboard-compose-argv.log")
}

// setDashboardAction sets hermetic interactive action env (+ optional dry-run + argv log).
func setDashboardAction(t *testing.T, req *Request, action string, dryRun bool) {
	t.Helper()
	req.ExtraEnv = append(req.ExtraEnv, envDashboardAction+"="+action)
	if dryRun {
		req.ExtraEnv = append(req.ExtraEnv, envDashboardDryRun+"=1")
	}
	logPath := dashComposeArgvLogPath(req)
	// Truncate so "not written" asserts are reliable.
	writeFile(t, logPath, "")
	req.ExtraEnv = append(req.ExtraEnv, envDashboardComposeArgvLog+"="+logPath)
}

func setDashboardToggles(req *Request, toggles string) {
	req.ExtraEnv = append(req.ExtraEnv, envDashboardToggles+"="+toggles)
}

// setupDashboardLinkedAhead creates main + linked feature branch with a commit ahead of main.
func setupDashboardLinkedAhead(t *testing.T, req *Request) string {
	t.Helper()
	linked := setupDashboardLinkedWorktree(t, req)
	commitAheadOnWorktree(t, linked, "feature-work.txt", "ahead of main for dashboard RUN\n")
	req.RepoDir = linked
	return linked
}

func readComposeArgvLog(t *testing.T, req *Request) string {
	t.Helper()
	return readFileEmptyOKDash(t, dashComposeArgvLogPath(req))
}

func readFileEmptyOKDash(t *testing.T, path string) string {
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

func composeArgvTokens(log string) []string {
	log = strings.TrimSpace(log)
	if log == "" {
		return nil
	}
	// Prefer one token per line; fall back to fields on a single line.
	if strings.Contains(log, "\n") {
		var toks []string
		for _, ln := range strings.Split(log, "\n") {
			ln = strings.TrimSpace(ln)
			if ln != "" {
				toks = append(toks, ln)
			}
		}
		return toks
	}
	return strings.Fields(log)
}

func argvHasToken(toks []string, want string) bool {
	for _, t := range toks {
		if t == want || strings.HasPrefix(t, want+"=") {
			return true
		}
	}
	// also allow "--agent-runner" "commandcode" as two tokens
	for i, t := range toks {
		if t == want && i+1 < len(toks) {
			return true
		}
	}
	return false
}

func argvHasAgentRunnerCommandcode(toks []string) bool {
	for i, t := range toks {
		if t == "--agent-runner=commandcode" {
			return true
		}
		if strings.HasPrefix(t, "--agent-runner=") && strings.Contains(t, "commandcode") {
			return true
		}
		if t == "--agent-runner" && i+1 < len(toks) && toks[i+1] == "commandcode" {
			return true
		}
	}
	return false
}

// assertComposeArgvRecipeDone checks default DONE recipe tokens (optional dry-run / add-all).
func assertComposeArgvRecipeDone(t *testing.T, req *Request, wantDryRun, wantAddAll bool) {
	t.Helper()
	raw := readComposeArgvLog(t, req)
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("WRK_DASHBOARD_COMPOSE_ARGV_LOG empty or missing; implementer must write composed argv before RUN compose (path=%s)",
			dashComposeArgvLogPath(req))
	}
	toks := composeArgvTokens(raw)
	need := []string{"--gen-commit-msg", "--commit", "--done", "--sync", "--tag-next", "--push", "--reinstall-local"}
	for _, n := range need {
		if !argvHasToken(toks, n) {
			t.Fatalf("compose argv missing %q; tokens=%v raw=%q", n, toks, raw)
		}
	}
	if !argvHasAgentRunnerCommandcode(toks) {
		t.Fatalf("compose argv missing --agent-runner=commandcode; tokens=%v", toks)
	}
	if argvHasToken(toks, "--merge-back") {
		t.Fatalf("DONE recipe must not include --merge-back; tokens=%v", toks)
	}
	if wantDryRun && !argvHasToken(toks, "--dry-run") {
		t.Fatalf("expected --dry-run in compose argv; tokens=%v", toks)
	}
	hasAddAll := argvHasToken(toks, "--add-all")
	if wantAddAll && !hasAddAll {
		t.Fatalf("dirty/enabled Add changes: expected --add-all; tokens=%v", toks)
	}
	if !wantAddAll && hasAddAll {
		t.Fatalf("disabled/clean Add changes: must not include --add-all; tokens=%v", toks)
	}
}

func assertComposeArgvRecipeMergeBack(t *testing.T, req *Request, wantDryRun bool) {
	t.Helper()
	raw := readComposeArgvLog(t, req)
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("WRK_DASHBOARD_COMPOSE_ARGV_LOG empty or missing (path=%s)", dashComposeArgvLogPath(req))
	}
	toks := composeArgvTokens(raw)
	need := []string{"--gen-commit-msg", "--commit", "--merge-back", "--sync", "--tag-next", "--push", "--reinstall-local"}
	for _, n := range need {
		if !argvHasToken(toks, n) {
			t.Fatalf("merge-back compose argv missing %q; tokens=%v raw=%q", n, toks, raw)
		}
	}
	if !argvHasAgentRunnerCommandcode(toks) {
		t.Fatalf("compose argv missing --agent-runner=commandcode; tokens=%v", toks)
	}
	if argvHasToken(toks, "--done") {
		t.Fatalf("MERGE BACK recipe must not include --done; tokens=%v", toks)
	}
	if wantDryRun && !argvHasToken(toks, "--dry-run") {
		t.Fatalf("expected --dry-run; tokens=%v", toks)
	}
}

func assertLinkedWorktreeStillPresent(t *testing.T, req *Request) {
	t.Helper()
	if req.WtDir == "" {
		t.Fatal("WtDir empty")
	}
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
}

// assertDryRunComposeEvidence requires that real compose dry-run path ran (not static snapshot only).
func assertDryRunComposeEvidence(t *testing.T, resp *Response, primary string) {
	t.Helper()
	all := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(all)
	// Static P2 snapshot alone is insufficient for RUN.
	if strings.Contains(resp.Stdout, "[•]") && !strings.Contains(lower, "would") &&
		!strings.Contains(all, "merge --ff-only") && !strings.Contains(lower, "dry-run") &&
		!strings.Contains(lower, "planned") {
		// still may fail below
	}
	hasPlan := strings.Contains(all, "merge --ff-only") ||
		strings.Contains(lower, "would:") ||
		strings.Contains(lower, "dry-run") ||
		strings.Contains(lower, "planned")
	if !hasPlan {
		t.Fatalf("RUN must invoke real multi-stage compose (dry-run plan evidence); stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if primary == "done" {
		// Prefer remove plan for --done dry-run
		if !strings.Contains(all, "worktree remove") && !strings.Contains(lower, "remove") &&
			!strings.Contains(all, "merge --ff-only") {
			t.Fatalf("done dry-run plan missing merge/remove evidence; out=%q err=%q", resp.Stdout, resp.Stderr)
		}
	}
	if primary == "merge-back" {
		// Must not remove worktree under merge-back dry-run when plan is visible
		if strings.Contains(all, "worktree remove") || strings.Contains(all, "worktree removed:") {
			t.Fatalf("merge-back dry-run must not plan worktree remove; out=%q err=%q", resp.Stdout, resp.Stderr)
		}
	}
}

// --- P3b tty-watch (real PTY Bubble Tea) ---

func lookPathTTYWatch(t *testing.T) string {
	t.Helper()
	if p, err := exec.LookPath("tty-watch"); err == nil && p != "" {
		return p
	}
	// Common install locations (do not skip if found here).
	candidates := []string{
		"/Users/xhd2015/go/bin/tty-watch",
		filepath.Join(os.Getenv("HOME"), "go", "bin", "tty-watch"),
		filepath.Join(os.Getenv("GOPATH"), "bin", "tty-watch"),
		"/usr/local/bin/tty-watch",
		"/opt/homebrew/bin/tty-watch",
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	t.Skip("tty-watch not found on PATH or common install paths; install tty-watch to run dashboard TTY leaves")
	return ""
}

func dashTTYWatchHome(req *Request) string {
	return filepath.Join(req.WorkRoot, "tty-watch-home")
}

func dashTTYSessionID(req *Request, leaf string) string {
	// Stable, unique-ish id per leaf under WorkRoot.
	base := filepath.Base(req.WorkRoot)
	if base == "" || base == "." {
		base = "dash"
	}
	return "dash-" + leaf + "-" + base
}

// runDashboardTTYWatch starts bare wrk under tty-watch, waits for dashboard UI,
// sends keys, waits for terminal exit, always kills the session.
// keys are passed to `tty-watch send` one at a time (use "q", "\r", etc.).
// After batch RUN the product re-opens the TUI (stay-in-TUI); leaves that only
// RUN should wait for argv log then send "q" (see runDashboardTTYWatchRunThenQuit).
func runDashboardTTYWatch(t *testing.T, req *Request, sessionLeaf string, extraEnv []string, keys ...string) (aliveSnap string, finalSnap string) {
	t.Helper()
	return runDashboardTTYWatchOpts(t, req, sessionLeaf, extraEnv, keys, false)
}

// runDashboardTTYWatchRunThenQuit sends keys, waits until compose argv log is
// non-empty (RUN completed), then sends "q" so the tea loop exits.
func runDashboardTTYWatchRunThenQuit(t *testing.T, req *Request, sessionLeaf string, extraEnv []string, keys ...string) (aliveSnap string, finalSnap string) {
	t.Helper()
	return runDashboardTTYWatchOpts(t, req, sessionLeaf, extraEnv, keys, true)
}

func runDashboardTTYWatchOpts(t *testing.T, req *Request, sessionLeaf string, extraEnv []string, keys []string, waitComposeThenQuit bool) (aliveSnap string, finalSnap string) {
	t.Helper()
	tw := lookPathTTYWatch(t)
	bin := getWrkBin(t)
	if req.RepoDir == "" {
		t.Fatal("runDashboardTTYWatch: RepoDir empty")
	}

	twHome := dashTTYWatchHome(req)
	if err := os.MkdirAll(twHome, 0o755); err != nil {
		t.Fatalf("mkdir TTY_WATCH_HOME: %v", err)
	}
	sid := dashTTYSessionID(req, sessionLeaf)

	// Ensure log path exists when dry-run RUN is used.
	logPath := dashComposeArgvLogPath(req)
	_ = os.WriteFile(logPath, []byte{}, 0o644)

	env := append([]string{}, os.Environ()...)
	// Drop conflicting keys then set isolation.
	env = filterEnvKeys(env, "WRK_HOME", "WRK_DATE", "TTY_WATCH_HOME", "TERM",
		envDashboardAction, envDashboardDryRun, envDashboardComposeArgvLog, envDashboardToggles)
	env = append(env,
		"WRK_HOME="+req.WrkHome,
		"WRK_DATE="+wrkDate,
		"TTY_WATCH_HOME="+twHome,
		// Avoid Bubble Tea OSC background-query hang on non-responding PTYs.
		"TERM=dumb",
	)
	for _, e := range extraEnv {
		env = append(env, e)
	}
	// Prefer argv log path when not already set.
	hasArgvLog := false
	for _, e := range extraEnv {
		if strings.HasPrefix(e, envDashboardComposeArgvLog+"=") {
			hasArgvLog = true
			break
		}
	}
	if !hasArgvLog {
		env = append(env, envDashboardComposeArgvLog+"="+logPath)
	}

	// Build: tty-watch run --detach --session-id SID -- env ... wrk
	// Pass env via wrapper so child inherits TERM=dumb and WRK_*.
	runArgs := []string{
		"run", "--detach", "--session-id", sid, "--",
		"env",
		"TERM=dumb",
		"WRK_HOME=" + req.WrkHome,
		"WRK_DATE=" + wrkDate,
		"TTY_WATCH_HOME=" + twHome,
	}
	for _, e := range extraEnv {
		runArgs = append(runArgs, e)
	}
	if !hasArgvLog {
		runArgs = append(runArgs, envDashboardComposeArgvLog+"="+logPath)
	}
	runArgs = append(runArgs, bin)

	cmd := exec.Command(tw, runArgs...)
	cmd.Dir = req.RepoDir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tty-watch run: %v\nout=%s", err, out)
	}

	defer func() {
		kill := exec.Command(tw, "kill", sid)
		kill.Env = env
		_ = kill.Run()
	}()

	// Poll snapshot until dashboard content appears (process still alive).
	deadline := time.Now().Add(15 * time.Second)
	for {
		aliveSnap = ttyWatchSnapshot(t, tw, env, sid)
		if snapshotLooksLikeDashboard(aliveSnap) && !strings.Contains(aliveSnap, "[Terminal exited]") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for live dashboard snapshot; last=%q", aliveSnap)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Send keys (if any). Brief settle so first frame is interactive.
	if len(keys) > 0 {
		time.Sleep(200 * time.Millisecond)
	}
	for _, k := range keys {
		send := exec.Command(tw, "send", sid, k)
		send.Env = env
		if out, err := send.CombinedOutput(); err != nil {
			t.Fatalf("tty-watch send %q: %v\nout=%s", k, err, out)
		}
		// Bubble Tea needs time between keys on slow CI / complex views.
		time.Sleep(120 * time.Millisecond)
	}

	if waitComposeThenQuit {
		// Stay-in-TUI: wait for compose argv log, then cancel to exit process.
		deadline = time.Now().Add(60 * time.Second)
		for {
			if strings.TrimSpace(readComposeArgvLog(t, req)) != "" {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("timeout waiting for compose argv log after RUN keys; snap=%q",
					ttyWatchSnapshot(t, tw, env, sid))
			}
			time.Sleep(100 * time.Millisecond)
		}
		send := exec.Command(tw, "send", sid, "q")
		send.Env = env
		if out, err := send.CombinedOutput(); err != nil {
			t.Fatalf("tty-watch send q after RUN: %v\nout=%s", err, out)
		}
	}

	// Poll until terminal exits (or timeout).
	if len(keys) > 0 || waitComposeThenQuit {
		deadline = time.Now().Add(60 * time.Second)
		for {
			finalSnap = ttyWatchSnapshot(t, tw, env, sid)
			if strings.Contains(finalSnap, "[Terminal exited]") {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("timeout waiting for terminal exit after keys; last=%q", finalSnap)
			}
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		finalSnap = aliveSnap
	}
	return aliveSnap, finalSnap
}

func ttyWatchSnapshot(t *testing.T, tw string, env []string, sid string) string {
	t.Helper()
	cmd := exec.Command(tw, "snapshot", sid)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Session may already be gone; return output anyway.
		return string(out)
	}
	return string(out)
}

func snapshotLooksLikeDashboard(snap string) bool {
	s := strings.ToLower(snap)
	// Prefer title; also accept stage chrome if title line was overdrawn.
	if strings.Contains(s, "dashboard") {
		return true
	}
	if strings.Contains(s, "gen-commit-msg") && (strings.Contains(s, "cancel") || strings.Contains(s, "run")) {
		return true
	}
	if strings.Contains(s, "add changes") && (strings.Contains(s, "merge back") || strings.Contains(s, "done")) {
		return true
	}
	return false
}

func filterEnvKeys(env []string, keys ...string) []string {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	var out []string
	for _, e := range env {
		eq := strings.IndexByte(e, '=')
		if eq <= 0 {
			out = append(out, e)
			continue
		}
		if drop[e[:eq]] {
			continue
		}
		out = append(out, e)
	}
	return out
}

func ensureDashboardHelpersUsed() {
	_ = readDashboardEvents
	_ = setupDashboardMainRepo
	_ = resolvePath
	_ = wantDashboardCreateWorktree
	_ = wantDashboardCreateWorktreeWithTask
	_ = worktreesRoot
	_ = listWorktreeEntries
	_ = assertNoWorktreesCreated
	_ = assertCreateOKPath
	_ = assertMutexNewMode
	_ = setupDashboardLinkedWorktree
	_ = markDashboardDirtyUntracked
	_ = lineContaining
	_ = stdoutHasAny
	_ = stdoutHasAllFold
	_ = assertDashboardSnapshotCore
	_ = assertAddChangesGlyph
	_ = assertMergeBackDefaultSelected
	_ = setDashboardAction
	_ = setDashboardToggles
	_ = setupDashboardLinkedAhead
	_ = readComposeArgvLog
	_ = readFileEmptyOKDash
	_ = composeArgvTokens
	_ = argvHasToken
	_ = argvHasAgentRunnerCommandcode
	_ = assertComposeArgvRecipeDone
	_ = assertComposeArgvRecipeMergeBack
	_ = assertLinkedWorktreeStillPresent
	_ = assertDryRunComposeEvidence
	_ = dashComposeArgvLogPath
	_ = lookPathTTYWatch
	_ = runDashboardTTYWatch
	_ = runDashboardTTYWatchRunThenQuit
	_ = dashTTYWatchHome
	_ = snapshotLooksLikeDashboard
	_ = filterEnvKeys
	_ = time.RFC3339
}
```
