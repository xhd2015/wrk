# Scenario

**Feature**: wrk --status Status line wording is `clean` or `dirty (… added…)`

```
# clean checkout root
clean myrepo -> wrk --status -> Status: clean

# untracked file maps to wrk "untracked" bucket
myrepo + ?? new.txt -> wrk --status -> dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)
```

## Preconditions

- Git must be available.
- In-process L2 via `wrkcli.Capture` (no product binary rebuild).
- Each leaf uses isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Descendants init a single-repo fixture, set `req.RepoDir` / `req.MainRepo`,
   and run `wrk --status` from the repo root (default `req.Args`).

## Context

- Asserts pin the full stdout block; the load-bearing line is `Status: …`.
- Dirty wording always includes all five buckets (zeros when empty).
- Porcelain `??` counts as **untracked**; `A`/`AM` as **staged** (path-once).
- Layer: **L2 in-process CLI** — not binary e2e.
- Stdout uses `assert.Output` v3 full-match templates (trailing `\n` required).

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
	"github.com/xhd2015/gitops/git/git_isolated"
)

const wrkDate = "2026-06-30"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if len(req.Args) == 0 {
		req.Args = []string{"--status"}
	}
	return nil
}

func statusFormatCaptureEnv(req *Request) []string {
	return []string{
		"WRK_HOME=" + req.WrkHome,
		"WRK_DATE=" + wrkDate,
	}
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
	return strings.TrimSpace(git_isolated.MustOutput(t, dir, args...))
}

// statusFormatInitRepo creates a clean main-branch repo with one README commit.
func statusFormatInitRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func statusFormatCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

// statusFormatBlockTemplate builds a v3 full-match template for a main-root
// status block with the given Status value (plain, no color).
func statusFormatBlockTemplate(t *testing.T, repoDir, statusLine string) string {
	t.Helper()
	body := fmt.Sprintf("Dir:          .\nBranch:       main\n%s\nStatus:       %s\nRemote:       (no upstream)\n",
		statusFormatCommitLine(t, repoDir), statusLine)
	// Escape metacharacters for v3 raw-regex content lines.
	escaped := make([]string, 0, 6)
	for _, line := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
		escaped = append(escaped, regexpQuoteMetaLine(line))
	}
	return "---\nversion: 3\n---\n" + strings.Join(escaped, "\n") + "\n"
}

// regexpQuoteMetaLine QuoteMeta's a full content line for assert v3 (raw regex).
func regexpQuoteMetaLine(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '.', '+', '*', '?', '(', ')', '[', ']', '{', '}', '|', '^', '$', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertStatusFormatOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}

// keep assert import used when leaves only call assert.Output via helpers
var _ = assert.Output
```
