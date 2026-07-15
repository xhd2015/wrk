# Scenario

**Feature**: detect task-like positionals when the user forgot `-t` / `--task`

```
# two-arg: wrk <dir> <arg2> without -t
# task-like arg2 -> TTY/WRK_TASK_LIKE_CONFIRM prompt Treat as --task? ; non-TTY error+hint; -y auto-promote
wrk <dir> "fix login" -> detect task-like -> promote or reject

# one-arg: wrk <arg1> without -t
# task-like arg1 that does not resolve as source -> same interactive / non-TTY rules; promote creates from cwd
wrk "fix login" -> detect task-like -> promote create from cwd with --task
```

## Preconditions

- Git available; isolated `{WRK_HOME}`; `WRK_DATE=2026-06-30`.
- Task-like when **any**: ASCII whitespace (after trim non-empty), **or** length > 120 bytes, **or** single path component would exceed 255 bytes / ENAMETOOLONG.
- **Never** task-like when path-like (`/`, `\`, leading `~`/`./`/`../`), resolves to an existing directory, or single-arg source resolve succeeds.
- Escape hatch: `WRK_TASK_LIKE_CONFIRM=1` + piped `StdinInput` treats stdin as interactive for the treat-as-task prompt (prefer over `UseScriptTTY`).
- Leaves pass task text via `TargetDir` / `SpawnDir` positionals — **not** `TaskDesc` (which injects `-t`/`--task`).

## Steps

- Shared helpers build a `myrepo` main checkout and assemble two-arg / one-arg invocations.
- Descendants set positionals, `-y` / confirm env / stdin, and assert promote vs target-dir vs error.

## Context

- Promote → default `{WRK_HOME}/worktrees/…` naming with slug from full task text; no spawn under prose path.
- `-t` already set → second positional remains target-dir; no treat-as-task prompt.

```go
import (
	"os"
	"path/filepath"
	"strings"
)

const (
	envTaskLikeConfirm = "WRK_TASK_LIKE_CONFIRM=1"
	taskLikeSpaces     = "fix the login bug"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	ensureForgotTaskFlagHelpersUsed()
	return nil
}

func initMyrepoForForgotTask(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	return mainRepo
}

// setupTwoArg: wrk <mainRepo> <arg2> from WorkRoot (shell cwd).
func setupTwoArg(t *testing.T, req *Request, arg2 string) {
	t.Helper()
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	req.SpawnDir = arg2
}

// setupOneArg: wrk <arg1> from inside mainRepo cwd.
func setupOneArg(t *testing.T, req *Request, arg1 string) {
	t.Helper()
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = mainRepo
	req.TargetDir = arg1
}

func wantPromotedWorktree(req *Request, task string) string {
	return worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify(task), 0)
}

func wantPromotedBranch(task string) string {
	return branchNameWithTask("main", wrkDate, slugify(task), 0)
}

func assertTaskLikeErrorTwoArg(t *testing.T, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for task-like second positional; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "task") {
		t.Fatalf("stderr should mention task description, got %q", resp.Stderr)
	}
	if !strings.Contains(low, "target") && !strings.Contains(low, "directory") {
		t.Fatalf("stderr should say not a target directory, got %q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "-t") && !strings.Contains(low, "task") {
		t.Fatalf("stderr should hint -t/--task, got %q", resp.Stderr)
	}
	// Prefer explicit -t in hint when present.
	if !strings.Contains(resp.Stderr, "-t") && !strings.Contains(resp.Stderr, "--task") {
		t.Fatalf("stderr should include -t or --task hint, got %q", resp.Stderr)
	}
	// stdout must not be a successful worktree path under WRK_HOME or a prose target.
	out := strings.TrimSpace(resp.Stdout)
	if out != "" && (strings.Contains(out, "worktrees") || filepath.IsAbs(out)) {
		// Allow empty only preferred; non-empty must not look like a created path with success semantics.
		// Hard rule: exit already non-zero; still reject if stdout is clearly a worktree path that was created.
		if _, statErr := os.Stat(out); statErr == nil {
			t.Fatalf("must not create worktree on task-like non-TTY error; stdout path exists: %q", out)
		}
	}
}

func assertTaskLikeErrorOneArg(t *testing.T, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for task-like first positional; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "task") {
		t.Fatalf("stderr should mention task description, got %q", resp.Stderr)
	}
	if !strings.Contains(low, "source") && !strings.Contains(low, "directory") {
		t.Fatalf("stderr should say not a source directory, got %q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "-t") && !strings.Contains(resp.Stderr, "--task") {
		t.Fatalf("stderr should include -t or --task hint, got %q", resp.Stderr)
	}
	out := strings.TrimSpace(resp.Stdout)
	if out != "" {
		if _, statErr := os.Stat(out); statErr == nil {
			t.Fatalf("must not create worktree on task-like non-TTY error; stdout path exists: %q", out)
		}
	}
}

func assertPromotedTaskCreate(t *testing.T, req *Request, resp *Response, err error, task string) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 after promote, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := wantPromotedWorktree(req, task)
	got := strings.TrimSpace(resp.Stdout)
	if got != want {
		t.Fatalf("stdout path: want promoted WRK_HOME worktree %q, got %q", want, got)
	}
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := wantPromotedBranch(task)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	// Must not spawn under prose path as fixed target-dir.
	prose := filepath.Join(req.WorkRoot, task)
	assertFileNotExists(t, prose)
}

func ensureForgotTaskFlagHelpersUsed() {
	_ = envTaskLikeConfirm
	_ = taskLikeSpaces
	_ = initMyrepoForForgotTask
	_ = setupTwoArg
	_ = setupOneArg
	_ = wantPromotedWorktree
	_ = wantPromotedBranch
	_ = assertTaskLikeErrorTwoArg
	_ = assertTaskLikeErrorOneArg
	_ = assertPromotedTaskCreate
}
```
