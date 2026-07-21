# Scenario

**Feature**: wrk skill --list|--show|--install for the single embedded wrk skill

```
# early dispatch — no git checkout required
wrk skill <flags> -> embedded SKILL.md (go:embed)

# flag actions (no skill name argument; Shape 1)
wrk skill --list | -l -> stdout wrk\n
wrk skill --show [--header] -> embedded SKILL.md or YAML header only
wrk skill --install [--cursor --dry-run] -> install.HandleInstall

# help / empty
wrk skill | --help | -h -> skill-level usage (exit 0)

# mutual exclusion + breaking change
wrk skill ... + other mode flag -> non-zero exit
wrk skill list|show|install (subcommand) -> non-zero exit
```

## Preconditions

- The wrk Go module is at `go-pkgs/cmd/wrk/` (three levels above this tree root).
- Go toolchain is available on PATH.
- Session wrk binary is built once per doctest run to
  process-local `MkdirTemp` wrk binary (in-memory mutex).
- Embedded skill content lives in `go-pkgs/wrkcli/SKILL.md` (compiled into the
  binary); doctest `show/basic` asserts `WRK_SKILL_DOCTEST_MARKER` and
  `name: wrk` in stdout.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Descendants set `req.RepoDir` plus `req.Args` for skill flag args after
   the leading `skill` token.

## Context

- Skill commands do not require a git repository; cwd is a neutral empty dir
  unless a descendant overrides `req.RepoDir`.
- Only `WRK_HOME` is passed via `skillWrkEnv`; no `WRK_SKILLS_ROOT`.
- Stdout assertions use `assert.Output` v2 full-match templates where output is
  bounded; `show/basic` uses substring checks for embedded content.
- User-facing help and list/show stdout must end with trailing `\n`.

```go
import (
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

const embeddedSkillMarker = "WRK_SKILL_DOCTEST_MARKER"

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

// Process-local wrk binary (one-process suite; in-memory mutex, not session flock).
var (
	wrkBinMu   sync.Mutex
	wrkBinPath string
	wrkBinErr  error
	// wrkModRoot set from d.DOCTEST_ROOT in root Setup.
	wrkModRoot string
)

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
		t.Fatal("wrkModRoot unset; root Setup must run first")
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
	if root := findModuleRoot(d.DOCTEST_ROOT); root != "" {
		wrkModRoot = root
	} else {
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
	ensureSkillHelpersUsed()
	return nil
}

func skillWrkEnv(req *Request) []string {
	return append(os.Environ(), "WRK_HOME="+req.WrkHome)
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", cwd, err)
	}
	return cwd
}

func skillHeaderStdoutV2() string {
	return v2StdoutTemplate(`---
name: wrk
description: >-
  Git worktree helper for isolated feature branches. Use when creating
  linked worktrees, merging back, checking status, linking deps, or
  looking up registered projects by basename.
---
`)
}

func assertEmbeddedSkillStdout(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, embeddedSkillMarker) {
		t.Fatalf("stdout missing embedded marker %q:\n%s", embeddedSkillMarker, stdout)
	}
	if !strings.Contains(stdout, "name: wrk") {
		t.Fatalf("stdout missing YAML name: wrk:\n%s", stdout)
	}
	if !strings.HasPrefix(stdout, "---\n") {
		t.Fatalf("stdout should start with YAML frontmatter, got:\n%s", stdout)
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout should end with trailing newline, got:\n%q", stdout)
	}
}

func assertSkillUsageStdout(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("expected exit 0 for skill help, got %d stderr=%q stdout=%q", exitCode, stderr, stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("stderr should be empty for skill help, got %q", stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected non-empty skill usage on stdout")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("skill usage stdout must end with trailing newline, got %q", stdout)
	}
	lower := strings.ToLower(stdout)
	for _, want := range []string{"--list", "--show", "--install"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("skill usage missing %q:\n%s", want, stdout)
		}
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

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func cursorSkillInstallDir(workRoot string) string {
	return filepath.Join(workRoot, ".cursor", "skills", "wrk")
}

func installDryRunCursorStdoutV2(t *testing.T, workRoot string) string {
	t.Helper()
	absWorkRoot, err := filepath.Abs(workRoot)
	if err != nil {
		t.Fatalf("abs work root: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(absWorkRoot); err == nil {
		absWorkRoot = resolved
	}
	skillDir := filepath.Join(absWorkRoot, ".cursor", "skills", "wrk")
	skillFile := filepath.Join(skillDir, "SKILL.md")
	body := fmt.Sprintf("\\[dry-run\\] Installed skill to: %s\n\\[dry-run\\]   create: %s\n",
		skillDir, skillFile)
	return v2StdoutTemplate(body)
}

func ensureSkillHelpersUsed() {
	_ = skillHeaderStdoutV2
	_ = assertEmbeddedSkillStdout
	_ = assertSkillUsageStdout
	_ = installDryRunCursorStdoutV2
	_ = cursorSkillInstallDir
}
```
