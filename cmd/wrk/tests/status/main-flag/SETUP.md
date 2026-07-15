# Scenario

**Feature**: wrk --main --status runs status of the checkout's main repo (no nested shell)

```
# composition: --main + --status → status of main repo, not shell
wrk --main --status from linked or main cwd
  -> ResolveMainRepo(ShowToplevel(workDir))
  -> runStatus(mainRepo, …) with Dir labels vs original invocation cwd

# flag order irrelevant
wrk --status --main  -> same as wrk --main --status

# event command is status (not main)
events.jsonl command="status"; args include --main and --status
```

## Preconditions

- Shared root harness: session `wrk` binary, isolated `WRK_HOME` / `WorkRoot`.
- Parent `status/SETUP.md` defaults `Args` to `["--status"]`; this node defaults to
  the composition `["--main", "--status"]`. Leaves may override.
- Do **not** redefine root `Request` / `Response` / `Run`.
- **Dir-aware equivalence**: same blocks and Branch/Commit/Status/Master/Remote as
  `wrk --status` from main; **Dir may differ** when invocation cwd ≠ main (rewrite
  reference Dir lines with `statusDirLine(invCwd, path)`).

## Steps

1. Descendants build a main checkout and optional external / in-tree linked worktrees.
2. Set `RepoDir` (process cwd) and `Args` for the composition under test.
3. Happy leaves compare against reference main status with Dir rewrite for inv cwd.

## Context

- Pure `wrk --main` (shell) and pure `wrk --status` (current checkout) stay unchanged.
- In-tree linked cwd + `--main --status` must use full main-repo status (not the
  linked-cwd scan-only shortcut of plain `--status` from that cwd).
- `--fetch` / `--color` / `-v` remain allowed with the pair; other modes stay exclusive.
- When already at main, Dir layout matches plain `--status` from main.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type mainFlagEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	// Override parent status default of pure --status.
	req.Args = []string{"--main", "--status"}
	ensureMainFlagHelpersUsed()
	return nil
}

func setMainStatusArgs(req *Request, args ...string) {
	req.Args = args
	req.TargetDir = ""
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// runWrkCapture runs wrk with full stdout/stderr capture (no TrimSpace).
func runWrkCapture(t *testing.T, req *Request, dir string, args ...string) *Response {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = wrkEnv(req)
	resp, err := captureCommandOutput(cmd, "")
	if err != nil {
		t.Fatalf("wrk %v in %s: %v", args, dir, err)
	}
	return resp
}

func runStatusFromMain(t *testing.T, req *Request) *Response {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo empty; cannot run reference --status")
	}
	return runWrkCapture(t, req, req.MainRepo, "--status")
}

func assertExitZeroEmptyStderr(t *testing.T, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}

// rewriteStatusDirLines replaces each "Dir:          …" line in order with
// statusDirLine(invCwd, repoPaths[i]) so reference stdout (from main cwd) can be
// compared against composition stdout from a different invocation cwd.
func rewriteStatusDirLines(t *testing.T, invCwd string, repoPaths []string, stdout string) string {
	t.Helper()
	const prefix = "Dir:          "
	var b strings.Builder
	pathIdx := 0
	for _, line := range strings.SplitAfter(stdout, "\n") {
		if strings.HasPrefix(line, prefix) {
			if pathIdx >= len(repoPaths) {
				t.Fatalf("more Dir lines than repoPaths (%d): paths=%v\nstdout:\n%s", len(repoPaths), repoPaths, stdout)
			}
			hasNL := strings.HasSuffix(line, "\n")
			line = prefix + statusDirLine(t, invCwd, repoPaths[pathIdx])
			if hasNL {
				line += "\n"
			}
			pathIdx++
		}
		b.WriteString(line)
	}
	if pathIdx != len(repoPaths) {
		t.Fatalf("Dir line count %d != len(repoPaths) %d\nstdout:\n%s", pathIdx, len(repoPaths), stdout)
	}
	return b.String()
}

// assertStdoutMainStatusDirAware: composition has same non-Dir fields as status-from-main;
// Dir lines equal statusDirLine(req.RepoDir, each path in block order).
func assertStdoutMainStatusDirAware(t *testing.T, req *Request, resp *Response, blockPaths ...string) {
	t.Helper()
	ref := runStatusFromMain(t, req)
	if ref.ExitCode != 0 {
		t.Fatalf("reference wrk --status from main exit %d stderr=%q", ref.ExitCode, ref.Stderr)
	}
	if resp.ExitCode != ref.ExitCode {
		t.Fatalf("exit code: composition=%d reference=%d", resp.ExitCode, ref.ExitCode)
	}
	want := rewriteStatusDirLines(t, req.RepoDir, blockPaths, ref.Stdout)
	if resp.Stdout != want {
		t.Fatalf("stdout Dir-aware mismatch vs main status content\n--- composition ---\n%s\n--- want (ref Dirs rewritten for cwd %s) ---\n%s\n--- reference (main cwd) ---\n%s",
			resp.Stdout, req.RepoDir, want, ref.Stdout)
	}
}

// assertStdoutEqualsMainStatus: when invCwd is main, Dir rewrite is identity for
// scan-relative paths; still rewrite so external Dir uses statusDirLine (not old abs).
// Root Request has MainRepo + WtDir only (no WtDir2); callers with extra blocks should
// pass paths explicitly via assertStdoutMainStatusDirAware.
func assertStdoutEqualsMainStatus(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo empty")
	}
	paths := []string{req.MainRepo}
	if req.WtDir != "" {
		paths = append(paths, req.WtDir)
	}
	assertStdoutMainStatusDirAware(t, req, resp, paths...)
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func addInTreeLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func setupExternalMainFlagFixture(t *testing.T, req *Request) (mainRepo, wtDir, branch string) {
	t.Helper()
	return setupWrkWorktreeFromMain(t, req)
}

func eventsJSONLPathMainFlag(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readMainFlagEvents(t *testing.T, wrkHome string) []mainFlagEvent {
	t.Helper()
	data, err := os.ReadFile(eventsJSONLPathMainFlag(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []mainFlagEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev mainFlagEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func assertLastEventCommandStatusWithMain(t *testing.T, wrkHome string, wantExit int) {
	t.Helper()
	events := readMainFlagEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	ev := events[len(events)-1]
	if ev.Command != "status" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "status", ev.Command, events)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
	hasMain, hasStatus := false, false
	for _, a := range ev.Args {
		if a == "--main" {
			hasMain = true
		}
		if a == "--status" {
			hasStatus = true
		}
	}
	if !hasMain || !hasStatus {
		t.Fatalf("event args should include --main and --status, got %v", ev.Args)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}

func ensureMainFlagHelpersUsed() {
	_ = setMainStatusArgs
	_ = resolvePath
	_ = statusDirLine
	_ = runWrkCapture
	_ = runStatusFromMain
	_ = assertExitZeroEmptyStderr
	_ = rewriteStatusDirLines
	_ = assertStdoutMainStatusDirAware
	_ = assertStdoutEqualsMainStatus
	_ = assertEmptyStdout
	_ = addInTreeLinkedWorktree
	_ = setupExternalMainFlagFixture
	_ = eventsJSONLPathMainFlag
	_ = readMainFlagEvents
	_ = assertLastEventCommandStatusWithMain
	_ = fmt.Sprintf
}
```