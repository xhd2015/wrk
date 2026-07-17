# Scenario

**Feature**: wrk --projects-dep-graph shows registered projects, modules, and cross edges

```
# exclusive mode; no git cwd required
WRK_HOME/projects.json
  -> BuildInventory (P1)
  -> format Projects + CrossEdges
  -> stdout human graph + footer
  -> stderr warning: for soft-skipped missing paths

# help
wrk -h -> documents --projects-dep-graph

# mutual exclusion
wrk --projects-dep-graph + --projects|--list -> non-zero exclusive error
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- The wrk Go module is found by walking ancestors of `DOCTEST_ROOT` for `go.mod`.
- Go toolchain is available on PATH.
- Session wrk binary is built once per doctest run to
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk` (file-locked).
- Git is **not** required (fixtures are plain directories with `go.mod`).
- Classic TDD: flag absent → RED until CLI implements the mode.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`, neutral
   non-git `RepoDir` (`workspace/`).
2. Descendants set `req.Args` and seed `projects.json` / module fixtures as needed.
3. Root `Run` executes the session-built `wrk` binary with `WRK_HOME` set.
4. Leaf `Assert` checks exit code, stdout graph (v2 templates), and stderr.

## Context

- **projects.json** schema: `{version, projects[{path, added_at, source}]}`.
- Display order of projects matches `storage.ListProjects` (sorted absolute paths).
- Module `dir` uses scan convention: `.` for root, relative slash path for nested.
- Cross-edge owner label is `filepath.Base(OwnerProject)`.
- Stdout assertions use `github.com/xhd2015/doctest/assert` v2 full-match templates.
- Shared only: session binary under fixture cache; per-leaf temp dirs stay isolated.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/assert"
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
	return filepath.Join(fixtureCacheBase(t), DOCTEST_SESSION_ID)
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
		modRoot := findModuleRoot(DOCTEST_ROOT)
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors")
		}
		cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
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
	depGraphEnsureHelpersUsed()
	return nil
}

func depGraphWrkEnv(req *Request) []string {
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

type projectsJSONEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type projectsJSONFile struct {
	Version  int                 `json:"version"`
	Projects []projectsJSONEntry `json:"projects"`
}

// writeProjectsJSON seeds WRK_HOME/projects.json with the given absolute paths.
// Paths need not exist (used for soft-skip scenarios).
func writeProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []projectsJSONEntry
	for _, p := range paths {
		projects = append(projects, projectsJSONEntry{
			Path:    p,
			AddedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := projectsJSONFile{Version: 1, Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		t.Fatalf("mkdir WRK_HOME: %v", err)
	}
	path := filepath.Join(wrkHome, "projects.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

// writeGoMod writes a minimal go.mod with optional require lines (path@version).
func writeGoMod(t *testing.T, dir, modulePath string, requires ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("module ")
	b.WriteString(modulePath)
	b.WriteString("\n\ngo 1.22\n")
	if len(requires) > 0 {
		b.WriteString("\nrequire (\n")
		for _, r := range requires {
			parts := strings.SplitN(r, "@", 2)
			if len(parts) != 2 {
				t.Fatalf("require %q must be path@version", r)
			}
			fmt.Fprintf(&b, "\t%s %s\n", parts[0], parts[1])
		}
		b.WriteString(")\n")
	}
	writeFile(t, filepath.Join(dir, "go.mod"), b.String())
}

// initRootAndSubModuleRepo creates root + nested sub/ modules (no git required).
func initRootAndSubModuleRepo(t *testing.T, path, rootModule string, rootRequires ...string) {
	t.Helper()
	mkdirAll(t, path)
	writeGoMod(t, path, rootModule, rootRequires...)
	writeFile(t, filepath.Join(path, "root.go"), "package rootmod\n")
	subDir := filepath.Join(path, "sub")
	mkdirAll(t, subDir)
	writeGoMod(t, subDir, rootModule+"/sub")
	writeFile(t, filepath.Join(subDir, "sub.go"), "package sub\n")
}

// initSingleModuleRepo creates a directory with only a root go.mod.
func initSingleModuleRepo(t *testing.T, path, modulePath string, requires ...string) {
	t.Helper()
	mkdirAll(t, path)
	writeGoMod(t, path, modulePath, requires...)
	writeFile(t, filepath.Join(path, "main.go"), "package main\n")
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

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func wordPlural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}

// graphFooter returns the summary line (no trailing newline).
func graphFooter(projects, modules, edges int) string {
	return fmt.Sprintf("%d %s  ·  %d %s  ·  %d %s",
		projects, wordPlural(projects, "project", "projects"),
		modules, wordPlural(modules, "module", "modules"),
		edges, wordPlural(edges, "cross-edge", "cross-edges"),
	)
}

func projectHeader(absPath string) string {
	return fmt.Sprintf("project %s  (%s)", filepath.Base(absPath), absPath)
}

func moduleLine(modulePath, dir string) string {
	return fmt.Sprintf("  module %s  dir=%s", modulePath, dir)
}

func crossEdgeLine(depPath, version, ownerAbs string) string {
	return fmt.Sprintf("    → %s@%s  \\[%s\\]", depPath, version, filepath.Base(ownerAbs))
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertMutualExclusion(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "mutually exclusive") {
		return
	}
	if strings.Contains(lower, "exclusive") {
		return
	}
	if strings.Contains(lower, "wrk:") && strings.Contains(lower, "error") {
		return
	}
	// Fallback strict-ish contains via v1 for clear failure message.
	assert.Output(t, resp.Stderr, `<contains>
exclusive
</contains>`)
}

func assertMissingPathWarning(t *testing.T, stderr, missingPath string) {
	t.Helper()
	if !strings.Contains(stderr, "warning:") {
		t.Fatalf("stderr must contain warning: prefix, got %q", stderr)
	}
	if !strings.Contains(stderr, missingPath) {
		t.Fatalf("stderr must mention missing path %q, got %q", missingPath, stderr)
	}
	// Prefer the DSN wording when present; still accept other soft-skip phrases.
	lower := strings.ToLower(stderr)
	if strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "missing") ||
		strings.Contains(lower, "skip") {
		return
	}
	t.Fatalf("stderr warning should describe missing/skip project path, got %q", stderr)
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

func depGraphEnsureHelpersUsed() {
	_ = writeProjectsJSON
	_ = initRootAndSubModuleRepo
	_ = initSingleModuleRepo
	_ = graphFooter
	_ = projectHeader
	_ = moduleLine
	_ = crossEdgeLine
	_ = assertMutualExclusion
	_ = assertMissingPathWarning
	_ = assertExitZero
	_ = v2StdoutTemplate
	_ = assertOutputExact
	_ = eventsJSONLPath
	_ = readEvents
}
```
