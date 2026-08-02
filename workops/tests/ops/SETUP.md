# Scenario

**Feature**: workops library Phase 1 ops over isolated git fixtures

```
# per-leaf temp WorkRoot + optional linked worktree / wrk home
Caller -> workops.<Op>(paths, DryRun?) -> structured result / plan
# no process cwd / HOME mutation; parallel-safe
```

## Preconditions

- Go module root is the wrk repo (`github.com/xhd2015/wrk`); tree lives at
  `workops/tests/ops/`.
- Git available on PATH (`git_isolated` helpers skip if missing).
- Product package `github.com/xhd2015/wrk/workops` is the SUT (Classic TDD:
  missing until implementer lands it → compile/link **RED**).
- No `os.Setenv` / `t.Setenv` / `os.Chdir` / `t.Chdir` in harness or Run.
- Session paths only via `d *session.Doctest` (`d.DOCTEST_ROOT`,
  `d.DOCTEST_CASE`, `d.DOCTEST_SESSION_ID`) — never `os.Getenv` of those names.

## Context

- Root Setup allocates `{WorkRoot}` via `t.TempDir()` (symlink-resolved) and
  `{WorkRoot}/.wrk` as injectable `WrkHome`.
- Process cwd is undetermined — use absolute paths from `t.TempDir` / helpers.
- Git fixtures: `git_isolated` init/commit/worktree/tag (hook-free).
- Linked worktrees created with `git worktree add -b` under WorkRoot (not wrk
  CLI) so L2 stays library-only.

## Steps

1. Touch session inject via `d` (parallel-safe; no env writes).
2. Create isolated WorkRoot + WrkHome directories.
3. Leaves fill `req.Op`, fixtures, and Checkout paths.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Require session inject; never read DOCTEST_* via os.Getenv.
	if d == nil || d.DOCTEST_SESSION_ID == "" {
		return fmt.Errorf("missing d.DOCTEST_SESSION_ID")
	}
	_ = d.DOCTEST_ROOT
	_ = d.DOCTEST_CASE

	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	workopsEnsureHelpersUsed()
	return nil
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

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func commitFile(t *testing.T, repo, relPath, content, subject string) string {
	t.Helper()
	writeFile(t, filepath.Join(repo, relPath), content)
	runGitIsolated(t, repo, "add", relPath)
	runGitIsolated(t, repo, "commit", "-m", subject)
	return gitOutputIsolated(t, repo, "rev-parse", "HEAD")
}

func createLightweightTag(t *testing.T, repo, name, ref string) {
	t.Helper()
	if ref == "" {
		ref = "HEAD"
	}
	runGitIsolated(t, repo, "tag", name, ref)
}

func tagRefExists(t *testing.T, repo, name string) bool {
	t.Helper()
	err := git_isolated.Command(repo, "rev-parse", "--verify", "refs/tags/"+name).Run()
	return err == nil
}

func shortHEAD(t *testing.T, repo string) string {
	t.Helper()
	return gitOutputIsolated(t, repo, "rev-parse", "--short=7", "HEAD")
}

func revParseHEAD(t *testing.T, repo string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", "HEAD"))
}

func revParseRef(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
}

func currentBranch(t *testing.T, repo string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", "--abbrev-ref", "HEAD"))
}

// seedMainRepo creates a minimal main repo with one commit on main.
func seedMainRepo(t *testing.T, req *Request, name string) string {
	t.Helper()
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, name)
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# seed\n", "init")
	repo = resolvePath(t, repo)
	req.MainRepo = repo
	return repo
}

// addLinkedWorktree adds a linked worktree of main at wtName under WorkRoot.
func addLinkedWorktree(t *testing.T, req *Request, mainRepo, wtName, branch string) string {
	t.Helper()
	wtPath := filepath.Join(req.WorkRoot, wtName)
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtPath)
	wtPath = resolvePath(t, wtPath)
	req.WtDir = wtPath
	req.WtBranch = branch
	return wtPath
}

// seedMainWithLinkedWorktree: main + one linked wt on branch (clean, same tip).
func seedMainWithLinkedWorktree(t *testing.T, req *Request) {
	t.Helper()
	main := seedMainRepo(t, req, "myrepo")
	addLinkedWorktree(t, req, main, "wt-feature", "feature-work")
}

// seedMainWithAheadWorktree: main + linked wt with an extra commit on feature branch.
func seedMainWithAheadWorktree(t *testing.T, req *Request) {
	t.Helper()
	seedMainWithLinkedWorktree(t, req)
	commitFile(t, req.WtDir, "feature.txt", "ahead\n", "feature-work ahead")
}

// seedRootBumpRepo: v0.0.1 at first commit; second commit changes README (tag plan → v0.0.2).
func seedRootBumpRepo(t *testing.T, req *Request) string {
	t.Helper()
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	commitFile(t, repo, "README.md", "# v1\n", "init")
	createLightweightTag(t, repo, "v0.0.1", "")
	commitFile(t, repo, "README.md", "# v2\n", "bump root")
	repo = resolvePath(t, repo)
	req.MainRepo = repo
	return repo
}

// seedMultiScopeBumpRepo: root v0.0.1 + sub/v0.2.3 at baseline; both scopes
// owned files change at HEAD → plan v0.0.2 and sub/v0.2.4 (tagscope multi-scope).
func seedMultiScopeBumpRepo(t *testing.T, req *Request) string {
	t.Helper()
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	writeFile(t, filepath.Join(repo, "README.md"), "# root\n")
	writeFile(t, filepath.Join(repo, "sub", "lib.go"), "package sub\n")
	runGitIsolated(t, repo, "add", "README.md", "sub/lib.go")
	runGitIsolated(t, repo, "commit", "-m", "init root and sub")
	createLightweightTag(t, repo, "v0.0.1", "")
	createLightweightTag(t, repo, "sub/v0.2.3", "")
	writeFile(t, filepath.Join(repo, "README.md"), "# root v2\n")
	writeFile(t, filepath.Join(repo, "sub", "lib.go"), "package sub // changed\n")
	runGitIsolated(t, repo, "add", "README.md", "sub/lib.go")
	runGitIsolated(t, repo, "commit", "-m", "bump root and sub")
	repo = resolvePath(t, repo)
	req.MainRepo = repo
	return repo
}

// seedPushRepoWithOrigin: main on main + bare origin with upstream tracking.
func seedPushRepoWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	bare := filepath.Join(req.WorkRoot, "origin.git")
	runGitIsolated(t, req.WorkRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	repo := seedMainRepo(t, req, "myrepo")
	runGitIsolated(t, repo, "remote", "add", "origin", bare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	req.OriginBare = resolvePath(t, bare)
}

// writeProjectsJSON writes a minimal projects.json under wrkHome.
func writeProjectsJSON(t *testing.T, wrkHome string, paths []string) {
	t.Helper()
	type project struct {
		Path    string `json:"path"`
		AddedAt string `json:"added_at"`
		Source  string `json:"source"`
	}
	type file struct {
		Version  int       `json:"version"`
		Projects []project `json:"projects"`
	}
	pf := file{Version: 1}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, p := range paths {
		pf.Projects = append(pf.Projects, project{
			Path:    resolvePath(t, p),
			AddedAt: now,
			Source:  "manual",
		})
	}
	data, err := json.Marshal(pf)
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	mkdirAll(t, wrkHome)
	if err := os.WriteFile(filepath.Join(wrkHome, "projects.json"), data, 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertPathEqual(t *testing.T, got, want string) {
	t.Helper()
	g := resolvePath(t, got)
	w := resolvePath(t, want)
	if g != w {
		t.Fatalf("path: got %q want %q", g, w)
	}
}

func assertPathNotEqual(t *testing.T, a, b string) {
	t.Helper()
	if resolvePath(t, a) == resolvePath(t, b) {
		t.Fatalf("paths should differ: both %q", a)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !st.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

func workopsEnsureHelpersUsed() {
	_ = seedMainRepo
	_ = addLinkedWorktree
	_ = seedMainWithLinkedWorktree
	_ = seedMainWithAheadWorktree
	_ = seedRootBumpRepo
	_ = seedMultiScopeBumpRepo
	_ = seedPushRepoWithOrigin
	_ = writeProjectsJSON
	_ = tagRefExists
	_ = shortHEAD
	_ = revParseHEAD
	_ = revParseRef
	_ = currentBranch
	_ = samePath
	_ = assertErrIsNil
	_ = assertPathEqual
	_ = assertPathNotEqual
	_ = assertDirExists
}
```
