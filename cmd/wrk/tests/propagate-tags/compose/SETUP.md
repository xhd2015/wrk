# Scenario

**Feature**: wrk --tag-next --propagate-tags runs tag-next then propagate-tags (P6 compose)

```
# compose pipeline (fixed order; flag argv free)
source (cwd) + registered consumers
  -> wrk --tag-next --propagate-tags [--push] [--dry-run]
  -> (1) tag-next plan/apply at source HEAD
  -> (2) push branch+tags + confirm when --push (non-dry-run)
  -> (3) propagate-tags using new/local tags (or planned next tags on dry-run)

# reject
wrk --tag-next --propagate-tags --json -> non-zero (json not valid with propagate-tags)

# bare still separate
wrk --propagate-tags alone does NOT auto tag-next  # parent tree
```

## Preconditions

- Nested root: **no inheritance** from parent `propagate-tags/` or monotree
  (`DOCTEST.md` firewall). Helpers are self-contained here.
- The wrk Go module is found by walking ancestors of `DOCTEST_ROOT` for `go.mod`.
- Go toolchain and git are available on PATH.
- Session wrk binary is built once per doctest run to
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk` (file-locked).
- Git helpers use `github.com/xhd2015/gitops/git/git_isolated` (hook-free).
- Apply leaves seed a local `file://` module proxy under `{WorkRoot}/modproxy` so
  `go mod tidy` can resolve the **next** release version offline.
- **P6 composition is not implemented yet** — Classic TDD RED (mutual exclusion).

## Target behavior (contract for implementer)

| Invocation | Expected |
|------------|----------|
| `--tag-next --propagate-tags` | Create next tag(s) on source; then bump consumers to those versions; build+commit consumers as bare propagate apply |
| `--tag-next --propagate-tags --dry-run` | Plan tag-next **and** propagate (using planned next versions); zero mutations |
| `--tag-next --propagate-tags --push` | Same as apply + push **branch + tags** with confirm line between stages; then propagate |
| `--tag-next --propagate-tags --json` | Hard error naming `--json` and `--propagate-tags` |
| `--propagate-tags` alone | Unchanged: uses existing tags only (no auto tag-next) |

Stdout stages separated by a blank line (major stage boundary), same family as
done-pipeline composition.

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Leaves build source/consumer fixtures (and optional bare origin), set
   `req.Args` / `req.RepoDir`, seed proxy when needed, capture snapshots.
3. Root `Run` executes session-built `wrk` with `WRK_HOME` (+ ExtraEnv).
4. Leaf `Assert` checks exit, multi-stage stdout, and side effects.

## Context

- Default fixture versions: source `example.com/lib` tagged `v1.0.0`, post-tag
  owned change → tag-next next `v1.0.1`; consumer requires `v1.0.0`.
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
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
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
	// Default version contract for compose root-bump leaves.
	req.OldTag = "v1.0.0"
	req.NextTag = "v1.0.1"
	req.ModulePath = "example.com/lib"
	composeEnsureHelpersUsed()
	return nil
}

func composeWrkEnv(req *Request) []string {
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

func createLightweightTag(t *testing.T, repo, name, ref string) {
	t.Helper()
	if ref == "" {
		ref = "HEAD"
	}
	runGitIsolated(t, repo, "tag", name, ref)
}

func listTags(t *testing.T, path string) string {
	t.Helper()
	out, err := git_isolated.Command(path, "tag", "--list").Output()
	if err != nil {
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

func commitNameOnly(t *testing.T, path string) string {
	t.Helper()
	return gitOutputIsolated(t, path, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
}

func tagRefExists(t *testing.T, repo, name string) bool {
	t.Helper()
	err := git_isolated.Command(repo, "rev-parse", "--verify", "refs/tags/"+name).Run()
	return err == nil
}

func remoteTagExists(t *testing.T, bareOrigin, name string) bool {
	t.Helper()
	out := gitOutputIsolated(t, bareOrigin, "show-ref", "--tags")
	prefix := "refs/tags/" + name
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == prefix {
			return true
		}
	}
	return false
}

func setupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func attachOriginAndPushMain(t *testing.T, repo, bareOrigin string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "add", "origin", bareOrigin)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
}

func goModPath(repo string) string {
	return filepath.Join(repo, "go.mod")
}

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

// setupComposeRootBump builds source lib + consumer app for tag-next→propagate.
// Source: tag OldTag at baseline, post-tag owned change → tag-next plans NextTag.
// App requires ModulePath@OldTag. Does not set Args (leaf does).
// withOrigin attaches bare origin and pushes main (for --push leaves).
func setupComposeRootBump(t *testing.T, req *Request, withOrigin bool) {
	t.Helper()
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	initGitRepo(t, libPath)
	writeGoMod(t, libPath, req.ModulePath, nil)
	writeFile(t, filepath.Join(libPath, "lib.go"),
		"package lib\n\nfunc Version() string { return \""+req.OldTag+"\" }\n")
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib "+req.OldTag)
	createLightweightTag(t, libPath, req.OldTag, "")
	// Post-tag owned change so tagscope plans next patch.
	writeFile(t, filepath.Join(libPath, "lib.go"),
		"package lib\n\nfunc Version() string { return \""+req.NextTag+"\" }\n")
	runGitIsolated(t, libPath, "add", "lib.go")
	runGitIsolated(t, libPath, "commit", "-m", "bump lib for next tag")
	libPath = resolvePath(t, libPath)

	if withOrigin {
		bare := setupBareOrigin(t, req.WorkRoot, "origin")
		attachOriginAndPushMain(t, libPath, bare)
		req.OriginBare = bare
	}

	initSingleModuleRepo(t, appPath, "example.com/app", []string{
		req.ModulePath + "@" + req.OldTag,
	})
	writeConsumerMainWithImports(t, appPath, req.ModulePath)
	appPath = resolvePath(t, appPath)

	// Proxy the *next* version (HEAD tree) for apply tidy after tag-next creates it.
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, req.ModulePath, req.NextTag, libPath)
	// Also seed old version so tidy/history is happy if needed.
	// Old tree is not at HEAD; seed a minimal zip from a temp snapshot of old content.
	seedOldModuleProxyFromContent(t, proxyRoot, req.ModulePath, req.OldTag,
		"package lib\n\nfunc Version() string { return \""+req.OldTag+"\" }\n")
	enableFileModuleProxy(t, req, proxyRoot)

	req.SourcePath = libPath
	req.AppPath = appPath
	writeProjectsJSON(t, req.WrkHome, libPath, appPath)
	req.RepoDir = libPath
	captureRepoSnapshots(t, req)
}

func seedOldModuleProxyFromContent(t *testing.T, proxyRoot, modulePath, version, libGo string) {
	t.Helper()
	tmp := filepath.Join(filepath.Dir(proxyRoot), "seed-old-"+version)
	mkdirAll(t, tmp)
	writeGoMod(t, tmp, modulePath, nil)
	writeFile(t, filepath.Join(tmp, "lib.go"), libGo)
	seedFileModuleProxy(t, proxyRoot, modulePath, version, tmp)
}

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

func seedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)

	modContent := readFile(t, filepath.Join(srcDir, "go.mod"))
	writeFile(t, filepath.Join(vDir, version+".mod"), modContent)
	writeFile(t, filepath.Join(vDir, version+".info"), fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version))
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
			base := filepath.Base(path)
			if path != srcDir && (base == ".git" || base == "sub") {
				return filepath.SkipDir
			}
			return nil
		}
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

func enableFileModuleProxy(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	abs, err := filepath.Abs(proxyRoot)
	if err != nil {
		t.Fatalf("abs proxy: %v", err)
	}
	proxyURL := "file://" + abs
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GONOSUMDB=*",
	)
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

// joinMajorStages joins multi-stage stdout with a blank line between stages.
func joinMajorStages(parts ...string) string {
	var cleaned []string
	for _, p := range parts {
		p = strings.TrimSuffix(p, "\n")
		if p == "" {
			continue
		}
		cleaned = append(cleaned, p)
	}
	return strings.Join(cleaned, "\n\n") + "\n"
}

// tagNextRootBumpPlanStdout is dry-run tag-next human plan for OldTag -> NextTag.
// Layout matches tagscope FormatPlanHuman (humanColTag=13, owned-changed reason width=31).
func tagNextRootBumpPlanStdout(oldTag, nextTag string) string {
	return tagNextDecisionLine(oldTag, nextTag) + "\n1 tag planned\n"
}

// tagNextRootBumpApplyStdout is apply tag-next human output including tagged line.
func tagNextRootBumpApplyStdout(oldTag, nextTag, short string) string {
	return tagNextDecisionLine(oldTag, nextTag) +
		fmt.Sprintf("\ntagged %s @ %s\n1 tag created\n", nextTag, short)
}

func tagNextDecisionLine(oldTag, nextTag string) string {
	// tagCol (13) + reasonCol (" owned changed" padded to 31) + " ->  " + next
	return padRight(oldTag, 13) + padRight(" owned changed", 31) + " ->  " + nextTag
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func sourceHeader(absPath string) string {
	return "source: " + absPath
}

func sourceReleaseLine(modulePath, version, tag string) string {
	return fmt.Sprintf("  %s  @ %s  (tag %s)", modulePath, version, tag)
}

func wouldUpdateHeader(consumerModule, projectBase string) string {
	return fmt.Sprintf("would: update %s  (project %s)", consumerModule, projectBase)
}

func updatedHeader(consumerModule, projectBase string) string {
	return fmt.Sprintf("updated %s  (project %s)", consumerModule, projectBase)
}

func versionBumpLine(depModule, oldVer, newVer string) string {
	return fmt.Sprintf("  %s  %s -> %s", depModule, oldVer, newVer)
}

func goBuildOkLine() string {
	return "  go build ./... ok"
}

func committedLine(short, subject string) string {
	return fmt.Sprintf("  committed %s  %s", short, subject)
}

func depsBumpSubject(depModule, version string) string {
	return fmt.Sprintf("chore(deps): bump %s to %s", depModule, version)
}

func wordPlural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}

func planFooter(modules, projects int) string {
	return fmt.Sprintf("would: update %d %s across %d %s",
		modules, wordPlural(modules, "module", "modules"),
		projects, wordPlural(projects, "project", "projects"),
	)
}

func applyFooter(modules, projects int) string {
	return fmt.Sprintf("updated %d %s across %d %s",
		modules, wordPlural(modules, "module", "modules"),
		projects, wordPlural(projects, "project", "projects"),
	)
}

// propStageDryRunStdout is the propagate-tags dry-run stage for compose (planned next release).
func propStageDryRunStdout(req *Request) string {
	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine(req.ModulePath, req.NextTag, req.NextTag))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(wouldUpdateHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine(req.ModulePath, req.OldTag, req.NextTag))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(planFooter(1, 1))
	b.WriteByte('\n')
	return b.String()
}

// propStageApplyStdout is the propagate-tags apply stage after tag-next created NextTag.
func propStageApplyStdout(req *Request, appShort, subject string) string {
	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine(req.ModulePath, req.NextTag, req.NextTag))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(updatedHeader("example.com/app", filepath.Base(req.AppPath)))
	b.WriteByte('\n')
	b.WriteString(versionBumpLine(req.ModulePath, req.OldTag, req.NextTag))
	b.WriteByte('\n')
	b.WriteString(goBuildOkLine())
	b.WriteByte('\n')
	b.WriteString(committedLine(appShort, subject))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(applyFooter(1, 1))
	b.WriteByte('\n')
	return b.String()
}

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

func assertSourceHEADUnchanged(t *testing.T, req *Request) {
	t.Helper()
	if got := headSHA(t, req.SourcePath); got != req.SourceHEADBefore {
		t.Fatalf("source HEAD mutated: before %s after %s", req.SourceHEADBefore, got)
	}
	if got := readFile(t, goModPath(req.SourcePath)); got != req.SourceGoModBefore {
		t.Fatalf("source go.mod mutated\nbefore:\n%s\nafter:\n%s", req.SourceGoModBefore, got)
	}
}

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

func assertGoModRequireVersion(t *testing.T, goMod, modulePath, version string) {
	t.Helper()
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "require ") {
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

func assertJSONRejectedWithPropagate(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "json") {
		t.Fatalf("stderr must mention json, got %q", resp.Stderr)
	}
	if !strings.Contains(lower, "propagate-tags") && !strings.Contains(lower, "propagate") {
		t.Fatalf("stderr must mention propagate-tags, got %q", resp.Stderr)
	}
	// Must not be a silent success into JSON tag-next-only path.
	if resp.Stdout != "" && strings.TrimSpace(resp.Stdout)[0] == '{' {
		t.Fatalf("stdout must not be JSON plan when compose rejects --json, got %q", resp.Stdout)
	}
}

func composeEnsureHelpersUsed() {
	_ = setupComposeRootBump
	_ = tagNextRootBumpPlanStdout
	_ = tagNextRootBumpApplyStdout
	_ = propStageDryRunStdout
	_ = propStageApplyStdout
	_ = joinMajorStages
	_ = assertDryRunNoMutation
	_ = assertAppDepsCommitted
	_ = assertGoModRequireVersion
	_ = assertJSONRejectedWithPropagate
	_ = tagRefExists
	_ = remoteTagExists
	_ = initNeutralCwd
	_ = v2StdoutTemplate
	_ = assertOutputExact
	_ = tagNextDecisionLine
	_ = padRight
}
```
