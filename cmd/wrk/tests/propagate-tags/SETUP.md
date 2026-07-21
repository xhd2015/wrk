# Scenario

**Feature**: wrk --propagate-tags plans and applies cross-project require bumps to source releases

```
# source main (cwd) + WRK_HOME projects.json consumers
git source tags
  -> ResolveSourceReleases(sourceMain)   # P1
  -> ListProjects other than source
  -> match require modules / version differs
  -> dry-run: stdout plan (would: update / drop replace); no mutation
  -> apply: drop replace + require + tidy
  -> P5: go build ./... per module; commit go.mod/go.sum if OK else warning:

# hard errors
non-git cwd | no numeric source tags | --list → non-zero

# dry-run host
wrk --dry-run alone → host list includes --propagate-tags
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- The wrk Go module is found by walking ancestors of `d.DOCTEST_ROOT` for `go.mod`.
- Go toolchain and git are available on PATH.
- Session wrk binary is built once per doctest run to
  process-local `MkdirTemp` wrk binary (in-memory mutex). (in-memory mutex, process-local).
- Git helpers use `github.com/xhd2015/gitops/git/git_isolated` (hook-free).
- Apply leaves may seed a local `file://` module proxy under `{WorkRoot}/modproxy`
  so `go mod tidy` can resolve synthetic `example.com/*` versions offline.
- P3 dry-run and P4 apply (edit+tidy) are implemented; P5 build+commit is Classic TDD RED.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Descendants build source/consumer git fixtures, seed `projects.json`, set
   `req.Args` / `req.RepoDir`, and capture pre-run snapshots for side effects.
3. Apply leaves may seed a module proxy and set `req.ExtraEnv` for GOPROXY.
4. Root `Run` executes the session-built `wrk` binary with `WRK_HOME` (+ ExtraEnv).
5. Leaf `Assert` checks exit code, stdout (v2 templates), stderr, and go.mod /
   tags / HEAD side effects as appropriate for plan vs apply.

## Context

- **projects.json** schema: `{version, projects[{path, added_at, source}]}`.
- Source cwd must be inside the source project's git work tree (main repo).
- Consumer projects are **other** registered paths; self is never an update target.
- Stdout assertions use `github.com/xhd2015/doctest/assert` v2 full-match templates.
- Shared only: session binary under fixture cache; per-leaf temp dirs stay isolated.

```go
import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/doctest/session"
	"sync"
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
	skipIfNoGit(t)
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
	propTagsEnsureHelpersUsed()
	return nil
}

func propTagsWrkEnv(req *Request) []string {
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func gitOutputIsolated(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git_isolated.MustOutput(t, dir, args...)
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

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	mkdirAll(t, cwd)
	return cwd
}

func initGitRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

// writeGoMod writes a minimal go.mod with optional require lines (path@version)
// and optional replace lines as "oldPath=>newPath" pairs.
func writeGoMod(t *testing.T, dir, modulePath string, requires []string, replaces ...string) {
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
	for _, repl := range replaces {
		parts := strings.SplitN(repl, "=>", 2)
		if len(parts) != 2 {
			t.Fatalf("replace %q must be old=>new", repl)
		}
		fmt.Fprintf(&b, "\nreplace %s => %s\n", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	writeFile(t, filepath.Join(dir, "go.mod"), b.String())
}

// initRootAndSubModuleRepo creates a git repo with root + nested sub/ modules.
// rootModule e.g. example.com/lib; sub module path is rootModule+"/sub".
func initRootAndSubModuleRepo(t *testing.T, path, rootModule string, rootRequires ...string) {
	t.Helper()
	initGitRepo(t, path)
	writeGoMod(t, path, rootModule, rootRequires)
	writeFile(t, filepath.Join(path, "root.go"), "package rootmod\n")
	subDir := filepath.Join(path, "sub")
	mkdirAll(t, subDir)
	writeGoMod(t, subDir, rootModule+"/sub", nil)
	writeFile(t, filepath.Join(subDir, "sub.go"), "package sub\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init root+sub modules")
}

// initSingleModuleRepo creates a git repo with only a root go.mod (+ optional requires/replaces).
func initSingleModuleRepo(t *testing.T, path, modulePath string, requires []string, replaces ...string) {
	t.Helper()
	initGitRepo(t, path)
	writeGoMod(t, path, modulePath, requires, replaces...)
	writeFile(t, filepath.Join(path, "main.go"), "package main\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init single module")
}

func tagRepo(t *testing.T, path, tag string) {
	t.Helper()
	runGitIsolated(t, path, "tag", tag)
}

func listTags(t *testing.T, path string) string {
	t.Helper()
	out, err := git_isolated.Command(path, "tag", "--list").Output()
	if err != nil {
		// empty tag list is fine
		return ""
	}
	return string(out)
}

func headSHA(t *testing.T, path string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, path, "rev-parse", "HEAD"))
}

func shortHEAD(t *testing.T, path string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, path, "rev-parse", "--short=7", "HEAD"))
}

func commitSubject(t *testing.T, path string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, path, "log", "-1", "--format=%s"))
}

func commitParentSHA(t *testing.T, path string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, path, "rev-parse", "HEAD^"))
}

// commitNameOnly lists paths changed in HEAD (no renames).
func commitNameOnly(t *testing.T, path string) string {
	t.Helper()
	return gitOutputIsolated(t, path, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
}

func goModPath(repo string) string {
	return filepath.Join(repo, "go.mod")
}

// captureRepoSnapshots fills Request pre-run fields for source and optional app.
func captureRepoSnapshots(t *testing.T, req *Request) {
	t.Helper()
	if req.SourcePath != "" {
		req.SourceGoModBefore = readFile(t, goModPath(req.SourcePath))
		req.SourceHEADBefore = headSHA(t, req.SourcePath)
		req.SourceTagsBefore = listTags(t, req.SourcePath)
	}
	if req.AppPath != "" {
		req.AppGoModBefore = readFile(t, goModPath(req.AppPath))
		req.AppHEADBefore = headSHA(t, req.AppPath)
		req.AppTagsBefore = listTags(t, req.AppPath)
	}
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

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func assertMutualExclusion(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "mutually exclusive") {
		return
	}
	if strings.Contains(lower, "exclusive") && strings.Contains(lower, "wrk:") {
		return
	}
	t.Fatalf("stderr should mention mutual exclusion, got %q", resp.Stderr)
}

// assertDryRunNoMutation checks go.mod / HEAD / tags unchanged for source and app.
func assertDryRunNoMutation(t *testing.T, req *Request) {
	t.Helper()
	if req.SourcePath != "" {
		if got := readFile(t, goModPath(req.SourcePath)); got != req.SourceGoModBefore {
			t.Fatalf("source go.mod mutated after dry-run\nbefore:\n%s\nafter:\n%s", req.SourceGoModBefore, got)
		}
		if got := headSHA(t, req.SourcePath); got != req.SourceHEADBefore {
			t.Fatalf("source HEAD mutated after dry-run: before %s after %s", req.SourceHEADBefore, got)
		}
		if got := listTags(t, req.SourcePath); got != req.SourceTagsBefore {
			t.Fatalf("source tags mutated after dry-run\nbefore:\n%s\nafter:\n%s", req.SourceTagsBefore, got)
		}
	}
	if req.AppPath != "" {
		if got := readFile(t, goModPath(req.AppPath)); got != req.AppGoModBefore {
			t.Fatalf("app go.mod mutated after dry-run\nbefore:\n%s\nafter:\n%s", req.AppGoModBefore, got)
		}
		if got := headSHA(t, req.AppPath); got != req.AppHEADBefore {
			t.Fatalf("app HEAD mutated after dry-run: before %s after %s", req.AppHEADBefore, got)
		}
		if got := listTags(t, req.AppPath); got != req.AppTagsBefore {
			t.Fatalf("app tags mutated after dry-run\nbefore:\n%s\nafter:\n%s", req.AppTagsBefore, got)
		}
	}
}

func wordPlural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}

// planFooter returns the dry-run summary line (no trailing newline).
func planFooter(modules, projects int) string {
	return fmt.Sprintf("would: update %d %s across %d %s",
		modules, wordPlural(modules, "module", "modules"),
		projects, wordPlural(projects, "project", "projects"),
	)
}

// applyFooter returns the apply summary line (no trailing newline).
func applyFooter(modules, projects int) string {
	return fmt.Sprintf("updated %d %s across %d %s",
		modules, wordPlural(modules, "module", "modules"),
		projects, wordPlural(projects, "project", "projects"),
	)
}

// sourceHeader returns the source: line.
func sourceHeader(absPath string) string {
	return "source: " + absPath
}

// sourceReleaseLine is one indented source release line.
func sourceReleaseLine(modulePath, version, tag string) string {
	return fmt.Sprintf("  %s  @ %s  (tag %s)", modulePath, version, tag)
}

// wouldUpdateHeader starts a consumer would-update block (dry-run).
func wouldUpdateHeader(consumerModule, projectBase string) string {
	return fmt.Sprintf("would: update %s  (project %s)", consumerModule, projectBase)
}

// updatedHeader starts a consumer updated block (apply).
func updatedHeader(consumerModule, projectBase string) string {
	return fmt.Sprintf("updated %s  (project %s)", consumerModule, projectBase)
}

// versionBumpLine is an indented require version arrow.
func versionBumpLine(depModule, oldVer, newVer string) string {
	return fmt.Sprintf("  %s  %s -> %s", depModule, oldVer, newVer)
}

// dropReplaceLine is a top-level would: drop replace action (dry-run).
func dropReplaceLine(depModule, projectBase string) string {
	return fmt.Sprintf("would: drop replace %s  (project %s)", depModule, projectBase)
}

// droppedReplaceLine is a top-level dropped replace action (apply).
func droppedReplaceLine(depModule, projectBase string) string {
	return fmt.Sprintf("dropped replace %s  (project %s)", depModule, projectBase)
}

// goBuildOkLine is the indented apply success line after bumps (P5).
func goBuildOkLine() string {
	return "  go build ./... ok"
}

// committedLine is the indented apply commit line (P5); short is --short=7.
func committedLine(short, subject string) string {
	return fmt.Sprintf("  committed %s  %s", short, subject)
}

// depsBumpSubject is the single-dep commit subject form (P5).
func depsBumpSubject(depModule, version string) string {
	return fmt.Sprintf("chore(deps): bump %s to %s", depModule, version)
}

// assertApplySourceUnchanged checks source go.mod / HEAD / tags unchanged.
func assertApplySourceUnchanged(t *testing.T, req *Request) {
	t.Helper()
	if req.SourcePath == "" {
		return
	}
	if got := readFile(t, goModPath(req.SourcePath)); got != req.SourceGoModBefore {
		t.Fatalf("source go.mod mutated after apply\nbefore:\n%s\nafter:\n%s", req.SourceGoModBefore, got)
	}
	if got := headSHA(t, req.SourcePath); got != req.SourceHEADBefore {
		t.Fatalf("source HEAD mutated after apply: before %s after %s", req.SourceHEADBefore, got)
	}
	if got := listTags(t, req.SourcePath); got != req.SourceTagsBefore {
		t.Fatalf("source tags mutated after apply\nbefore:\n%s\nafter:\n%s", req.SourceTagsBefore, got)
	}
}

// assertAppHEADUnchanged checks consumer HEAD and tags did not advance (no commit).
func assertAppHEADUnchanged(t *testing.T, req *Request) {
	t.Helper()
	if req.AppPath == "" {
		return
	}
	if got := headSHA(t, req.AppPath); got != req.AppHEADBefore {
		t.Fatalf("app HEAD mutated after apply (must not commit): before %s after %s", req.AppHEADBefore, got)
	}
	if got := listTags(t, req.AppPath); got != req.AppTagsBefore {
		t.Fatalf("app tags mutated after apply\nbefore:\n%s\nafter:\n%s", req.AppTagsBefore, got)
	}
}

// assertApplyNoGitMutation checks source go.mod/tags/HEAD and consumer tags/HEAD
// unchanged after apply (no commit path: build-fail or pre-P5). Consumer go.mod
// may still change when bumps apply without commit.
func assertApplyNoGitMutation(t *testing.T, req *Request) {
	t.Helper()
	assertApplySourceUnchanged(t, req)
	assertAppHEADUnchanged(t, req)
}

// assertAppDepsCommitted checks consumer got exactly one new commit whose parent
// is AppHEADBefore, subject matches, and only go.mod/go.sum paths are in the tree.
func assertAppDepsCommitted(t *testing.T, req *Request, wantSubject string) {
	t.Helper()
	if req.AppPath == "" {
		t.Fatal("assertAppDepsCommitted: AppPath empty")
	}
	gotHEAD := headSHA(t, req.AppPath)
	if gotHEAD == req.AppHEADBefore {
		t.Fatalf("app HEAD did not advance after successful build+commit (still %s)", gotHEAD)
	}
	parent := commitParentSHA(t, req.AppPath)
	if parent != req.AppHEADBefore {
		t.Fatalf("app commit parent want %s got %s (HEAD %s)", req.AppHEADBefore, parent, gotHEAD)
	}
	if got := listTags(t, req.AppPath); got != req.AppTagsBefore {
		t.Fatalf("app tags mutated after apply\nbefore:\n%s\nafter:\n%s", req.AppTagsBefore, got)
	}
	gotSubject := commitSubject(t, req.AppPath)
	if gotSubject != wantSubject {
		t.Fatalf("commit subject want %q got %q", wantSubject, gotSubject)
	}
	// Only go.mod / go.sum under edited module dirs (root module → those basenames).
	names := strings.Fields(strings.ReplaceAll(commitNameOnly(t, req.AppPath), "\n", " "))
	if len(names) == 0 {
		t.Fatal("commit has no file paths")
	}
	for _, n := range names {
		base := filepath.Base(n)
		if base != "go.mod" && base != "go.sum" {
			t.Fatalf("commit must only include go.mod/go.sum, got path %q (all: %v)", n, names)
		}
	}
}

// assertGoModRequireVersion checks go.mod text has require path at version
// (tolerates single-line or parenthesized require blocks).
func assertGoModRequireVersion(t *testing.T, goMod, modulePath, version string) {
	t.Helper()
	// go mod edit / tidy may reformat; look for path and version on same line.
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "require ") {
			// single-line: require path version
			fields := strings.Fields(trim)
			if len(fields) >= 3 && fields[1] == modulePath && fields[2] == version {
				return
			}
			continue
		}
		fields := strings.Fields(trim)
		if len(fields) >= 2 && fields[0] == modulePath && fields[1] == version {
			return
		}
	}
	t.Fatalf("go.mod missing require %s %s\n%s", modulePath, version, goMod)
}

// assertGoModNoReplace checks no replace directive for modulePath remains.
func assertGoModNoReplace(t *testing.T, goMod, modulePath string) {
	t.Helper()
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.Contains(trim, modulePath) {
			continue
		}
		if strings.HasPrefix(trim, "replace ") || strings.HasPrefix(trim, modulePath+" =>") ||
			strings.Contains(trim, "=>") && strings.HasPrefix(trim, modulePath) {
			t.Fatalf("go.mod still has replace for %s: %q\nfull:\n%s", modulePath, trim, goMod)
		}
	}
}

// writeConsumerMainWithImports writes main.go that blank-imports each module path
// so go mod tidy keeps those requires.
func writeConsumerMainWithImports(t *testing.T, dir string, importPaths ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("package main\n\n")
	if len(importPaths) > 0 {
		b.WriteString("import (\n")
		for _, p := range importPaths {
			fmt.Fprintf(&b, "\t_ %q\n", p)
		}
		b.WriteString(")\n\n")
	}
	b.WriteString("func main() {}\n")
	writeFile(t, filepath.Join(dir, "main.go"), b.String())
}

// writeConsumerMainBuildBreak writes main.go that imports modPath but references a
// missing exported name so go mod tidy succeeds and go build ./... fails (P5 gate).
func writeConsumerMainBuildBreak(t *testing.T, dir, modPath string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("package main\n\n")
	fmt.Fprintf(&b, "import lib %q\n\n", modPath)
	b.WriteString("func main() {\n")
	// Identifier guaranteed absent from fixture lib packages.
	b.WriteString("\t_ = lib.DoesNotExist\n")
	b.WriteString("}\n")
	writeFile(t, filepath.Join(dir, "main.go"), b.String())
}

// seedFileModuleProxy publishes srcDir as modulePath@version under a file://
// module proxy root so offline go mod tidy can resolve the version.
func seedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)

	modContent := readFile(t, filepath.Join(srcDir, "go.mod"))
	writeFile(t, filepath.Join(vDir, version+".mod"), modContent)
	writeFile(t, filepath.Join(vDir, version+".info"), fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version))
	// Append version to list if missing.
	listPath := filepath.Join(vDir, "list")
	existing := ""
	if data, err := os.ReadFile(listPath); err == nil {
		existing = string(data)
	}
	if !strings.Contains(existing, version) {
		writeFile(t, listPath, existing+version+"\n")
	}

	zipPath := filepath.Join(vDir, version+".zip")
	if err := writeModuleZip(zipPath, modulePath, version, srcDir); err != nil {
		t.Fatalf("write module zip %s: %v", zipPath, err)
	}
}

func writeModuleZip(zipPath, modulePath, version, srcDir string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	prefix := modulePath + "@" + version + "/"
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip nested module trees and VCS metadata inside the package root.
			base := filepath.Base(path)
			if path != srcDir && (base == ".git" || base == "sub") {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip .git files if any.
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || rel == ".git" {
			return nil
		}
		rel = filepath.ToSlash(rel)
		w, err := zw.Create(prefix + rel)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, in)
		_ = in.Close()
		return copyErr
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

// enableFileModuleProxy sets ExtraEnv so wrk's child `go` commands resolve
// modules via the local file proxy (offline-friendly tidy).
func enableFileModuleProxy(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	abs, err := filepath.Abs(proxyRoot)
	if err != nil {
		t.Fatalf("abs proxy: %v", err)
	}
	// file URLs need three slashes: file:///abs/path
	proxyURL := "file://" + abs
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GONOSUMDB=*",
	)
}

// propTagsEnsureHelpersUsed references helpers so unused-import / deadcode tools
// do not strip them before leaf packages link the full harness.
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

func propTagsEnsureHelpersUsed() {
	_ = sourceHeader
	_ = sourceReleaseLine
	_ = wouldUpdateHeader
	_ = updatedHeader
	_ = versionBumpLine
	_ = dropReplaceLine
	_ = droppedReplaceLine
	_ = goBuildOkLine
	_ = committedLine
	_ = depsBumpSubject
	_ = planFooter
	_ = applyFooter
	_ = assertDryRunNoMutation
	_ = assertApplyNoGitMutation
	_ = assertApplySourceUnchanged
	_ = assertAppHEADUnchanged
	_ = assertAppDepsCommitted
	_ = shortHEAD
	_ = commitSubject
	_ = commitParentSHA
	_ = commitNameOnly
	_ = assertGoModRequireVersion
	_ = assertGoModNoReplace
	_ = writeConsumerMainWithImports
	_ = writeConsumerMainBuildBreak
	_ = seedFileModuleProxy
	_ = enableFileModuleProxy
	_ = assertMutualExclusion
	_ = captureRepoSnapshots
	_ = writeProjectsJSON
	_ = initRootAndSubModuleRepo
	_ = initSingleModuleRepo
	_ = tagRepo
	_ = initNeutralCwd
	_ = v2StdoutTemplate
	_ = assertOutputExact
	_ = eventsJSONLPath
	_ = readEvents
}
```

