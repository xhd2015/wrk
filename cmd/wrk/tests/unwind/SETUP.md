# Scenario

**Feature**: wrk --unwind dry-run + apply peel/pin over checkout stack DAG

```
# isolated WRK_HOME + multi-repo stack under consumer checkout
stack (root + nested external/*) + go require edges
  -> wrk --unwind --dry-run [+ ship/land + gen-commit flags]
  -> cycle? Error mentioning cycle (no mutations)
  -> missing pin flags with edges? Error naming --tag-next/--push
  -> else would: peel <display-path>… free-first among dirty pending
  -> gen-commit: would: git add -A and/or leave-N as flags/porcelain
  -> apply (no --dry-run): peel free-first + pin; banner uses display-path
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- The wrk Go module is found by walking ancestors of `d.DOCTEST_ROOT` for `go.mod`.
- Go toolchain and **git** are available on PATH.
- Session wrk binary is built once per doctest run to
  `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk` (file-locked) for fixture
  create / worktree helpers when needed.
- Apply fixtures use local bare remotes + optional `file://` module proxy (no network).
- New display/add-all/leave-N contracts are Classic TDD (**RED** until implementer lands).

## Steps

1. Root `Setup` creates isolated `{WorkRoot}` and `{WorkRoot}/.wrk`.
2. Leaves seed stack fixtures (single main, 3-repo chain, 2-repo apply, or 2-cycle)
   and set `req.Args` / `req.RepoDir` / `req.PeelOrder` (display paths) / apply fields.
3. Root `Run` invokes wrk via `wrkcli.Capture` when `InProcess` (default for leaves).

## Context

- Peel **display paths** (not bare MainRepo basenames): relative checkout path vs
  `RepoDir` / invocation cwd — e.g. `external/dot-pkgs-main-2026-06-30`, `.` for
  primary at cwd. Assert via `would: peel <display-path>` and apply banner.
- DAG identity uses MainRepo abs; human pin short names remain basenames.
- Dirtiness v1: untracked file `DIRTY` under a stack checkout path (counts as N=1
  not-fully-staged for leave-N when not staged).
- Apply leaf fixtures also commit **ahead-of-main** work on linked WTs so land
  has content; pin leaves seed `{WorkRoot}/modproxy` + `ExtraEnv` for tidy.
- After materializing nested externals, consumer porcelain is cleaned before
  intentional dirtify so dirty-filter tests stay intentional.
- L2: leave `req.InProcess = true` on every leaf.

```go
import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	wrkDate = "2026-06-30"

	unwindRootModule     = "example.com/root"
	unwindAgentProModule = "example.com/agent-pro"
	unwindDotPkgsModule  = "example.com/dot-pkgs"
	unwindCycleAModule   = "example.com/cycle-a"
	unwindCycleBModule   = "example.com/cycle-b"

	labelRoot     = "root"
	labelAgentPro = "agent-pro"
	labelDotPkgs  = "dot-pkgs"
	labelCycleA   = "cycle-a"
	labelCycleB   = "cycle-b"

	// Apply pin baseline / next tags (root-bump style on leaf module).
	unwindApplyOldTag  = "v0.0.1"
	unwindApplyNextTag = "v0.0.2"
)

// harnessDoctest holds inject fields from d (no os.Setenv — Parallel-safe).
var (
	harnessMu          sync.Mutex
	harnessSessionID   string
	harnessDoctestRoot string
)

func adoptDoctestContext(d *session.Doctest) {
	if d == nil {
		return
	}
	harnessMu.Lock()
	defer harnessMu.Unlock()
	if d.DOCTEST_SESSION_ID != "" {
		harnessSessionID = d.DOCTEST_SESSION_ID
	}
	if d.DOCTEST_ROOT != "" {
		harnessDoctestRoot = d.DOCTEST_ROOT
	}
}

func doctestSessionID(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	sid := harnessSessionID
	harnessMu.Unlock()
	if sid == "" {
		t.Fatal("d.DOCTEST_SESSION_ID not set (expected adoptDoctestContext from Setup)")
	}
	return sid
}

func doctestRootPath(t *testing.T) string {
	t.Helper()
	harnessMu.Lock()
	root := harnessDoctestRoot
	harnessMu.Unlock()
	if root == "" {
		t.Fatal("d.DOCTEST_ROOT not set (expected adoptDoctestContext from Setup)")
	}
	return root
}

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
	return filepath.Join(fixtureCacheBase(t), doctestSessionID(t))
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
		modRoot := findModuleRoot(doctestRootPath(t))
		if modRoot == "" {
			t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
		}
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	adoptDoctestContext(d)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
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
	unwindEnsureHelpersUsed()
	return nil
}

func unwindWrkEnv(req *Request) []string {
	env := append(os.Environ(),
		"WRK_HOME="+req.WrkHome,
		"WRK_DATE="+wrkDate,
	)
	if len(req.ExtraEnv) > 0 {
		env = append(env, req.ExtraEnv...)
	}
	return env
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

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func gitOutputIsolated(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return git_isolated.MustOutput(t, dir, args...)
}

func revParseHEAD(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, dir, "rev-parse", "HEAD"))
}

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	if err := git_isolated.Init(path, "main"); err != nil {
		t.Fatalf("git init %s: %v", path, err)
	}
	runGitIsolated(t, path, "config", "user.email", git_isolated.DefaultUserEmail)
	runGitIsolated(t, path, "config", "user.name", git_isolated.DefaultUserName)
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", "init")
}

func writeGoModRequire(t *testing.T, dir, modulePath string, requires ...string) {
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

func branchNameMainDate() string {
	return "main-" + wrkDate
}

func runWrkBin(t *testing.T, req *Request, dir string, args ...string) string {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = unwindWrkEnv(req)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk %v exit %d stderr=%q", args, ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func markDirty(t *testing.T, checkout string) {
	t.Helper()
	writeFile(t, filepath.Join(checkout, "DIRTY"), "dirty\n")
}

func markCleanTracked(t *testing.T, checkout string) {
	t.Helper()
	// Ensure no leftover DIRTY from prior steps.
	_ = os.Remove(filepath.Join(checkout, "DIRTY"))
}

func commitAllAllowEmpty(t *testing.T, dir, msg string) {
	t.Helper()
	runGitIsolated(t, dir, "add", "-A")
	runGitIsolated(t, dir, "commit", "-m", msg, "--allow-empty")
}

// setupSingleMainDirty: main-only repo basename "root", dirty, no externals.
// RepoDir = main. Peel display = "." (checkout is cwd). Already on main → no land/pin flags needed.
func setupSingleMainDirty(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "main.go"), "package main\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "main.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "add module")
	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	markDirty(t, mainRepo)
	req.PeelOrder = []string{"."}
}

// peelDisplay returns statusDirLine-style relative path of checkout vs req.RepoDir.
// Mirrors wrkcli.statusDirLine: slash form; abs if Rel fails or leading ".." > 2.
func peelDisplay(t *testing.T, req *Request, checkout string) string {
	t.Helper()
	if req.RepoDir == "" {
		t.Fatal("peelDisplay: RepoDir must be set")
	}
	base := resolvePath(t, req.RepoDir)
	target := resolvePath(t, checkout)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	rel = filepath.Clean(rel)
	slash := filepath.ToSlash(rel)
	leading := 0
	for _, p := range strings.Split(slash, "/") {
		if p == ".." {
			leading++
			continue
		}
		break
	}
	if leading > 2 {
		return target
	}
	return slash
}

// setPeelOrderDisplays fills PeelOrder from free-first checkout paths relative to RepoDir.
func setPeelOrderDisplays(t *testing.T, req *Request, checkouts ...string) {
	t.Helper()
	order := make([]string, 0, len(checkouts))
	for _, c := range checkouts {
		order = append(order, peelDisplay(t, req, c))
	}
	req.PeelOrder = order
}

// stageAll stages all paths in checkout (git add -A) for fully-staged leave-N fixtures.
func stageAll(t *testing.T, checkout string) {
	t.Helper()
	runGitIsolated(t, checkout, "add", "-A")
}

// leaveLine builds the locked leave-N dry-run vocabulary for N not-fully-staged paths.
func leaveLine(n int) string {
	if n == 1 {
		return "would: leave 1 file uncommitted (use --add-all if necessary)"
	}
	return fmt.Sprintf("would: leave %d files uncommitted (use --add-all if necessary)", n)
}

func assertContainsInOrder(t *testing.T, s string, parts ...string) {
	t.Helper()
	at := 0
	for _, p := range parts {
		i := strings.Index(s[at:], p)
		if i < 0 {
			t.Fatalf("missing %q (in order) in:\n%s", p, s)
		}
		at += i + len(p)
	}
}

// setupThreeRepoChain creates:
//
//	root main + linked wt (consumer)
//	agent-pro main + external wt under consumer/external/
//	dot-pkgs main + external wt under consumer/external/
//
// go.mod edges: root requires agent-pro; agent-pro requires dot-pkgs.
// Caller dirties selected checkouts via markDirty and sets Args.
func setupThreeRepoChain(t *testing.T, req *Request) {
	t.Helper()

	// --- leaf: dot-pkgs ---
	dotMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, dotMain)
	writeGoModRequire(t, dotMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(dotMain, "pkg.go"), "package dotpkgs\n")
	runGitIsolated(t, dotMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, dotMain, "commit", "-m", "add dot-pkgs module")
	dotMain = resolvePath(t, dotMain)
	req.SecondRepo = dotMain

	// --- mid: agent-pro requires dot-pkgs ---
	agentMain := filepath.Join(req.WorkRoot, labelAgentPro)
	initGitRepoOnMain(t, agentMain)
	writeGoModRequire(t, agentMain, unwindAgentProModule, unwindDotPkgsModule+"@v0.0.0")
	writeFile(t, filepath.Join(agentMain, "agent.go"), "package agentpro\n")
	runGitIsolated(t, agentMain, "add", "go.mod", "agent.go")
	runGitIsolated(t, agentMain, "commit", "-m", "add agent-pro module")
	agentMain = resolvePath(t, agentMain)
	req.DepPath = agentMain

	// --- root consumer requires agent-pro ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindAgentProModule+"@v0.0.0")
	writeFile(t, filepath.Join(rootMain, "main.go"), "package main\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "main.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root module")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// linked worktree for root (linked → land flag needed when shipping)
	wtDir := runWrkBin(t, req, rootMain, "--new")
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// nest dep worktrees under consumer checkout (stack external members)
	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)

	agentExt := filepath.Join(extDir, labelAgentPro+"-"+branchNameMainDate())
	runGitIsolated(t, agentMain, "worktree", "add", "-b", branchNameMainDate(), agentExt)
	agentExt = resolvePath(t, agentExt)
	req.ExternalWtDir = agentExt

	dotExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, dotMain, "worktree", "add", "-b", branchNameMainDate(), dotExt)
	dotExt = resolvePath(t, dotExt)
	req.DepsLinkedWtDir = dotExt

	// Keep require edges on the *worktree* checkouts (same commits as mains for now).
	// Clean consumer porcelain after external/ appears.
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	commitAllAllowEmpty(t, wtDir, "ignore external stack members")

	req.RepoDir = wtDir
}

// dirtyAllThree marks root wt + agent-pro ext + dot-pkgs ext dirty.
func dirtyAllThree(t *testing.T, req *Request) {
	t.Helper()
	markDirty(t, req.WtDir)
	markDirty(t, req.ExternalWtDir)
	markDirty(t, req.DepsLinkedWtDir)
}

// dirtyMidAndRoot leaves leaf clean; mid + root dirty.
func dirtyMidAndRoot(t *testing.T, req *Request) {
	t.Helper()
	markCleanTracked(t, req.DepsLinkedWtDir)
	markDirty(t, req.ExternalWtDir)
	markDirty(t, req.WtDir)
}

// setupTwoCycleStack: consumer wt hosts external cycle-a and cycle-b with
// mutual requires (A requires B, B requires A). Both dirty.
func setupTwoCycleStack(t *testing.T, req *Request) {
	t.Helper()

	aMain := filepath.Join(req.WorkRoot, labelCycleA)
	initGitRepoOnMain(t, aMain)
	writeGoModRequire(t, aMain, unwindCycleAModule, unwindCycleBModule+"@v0.0.0")
	writeFile(t, filepath.Join(aMain, "a.go"), "package cyclea\n")
	runGitIsolated(t, aMain, "add", "go.mod", "a.go")
	runGitIsolated(t, aMain, "commit", "-m", "add cycle-a")
	aMain = resolvePath(t, aMain)
	req.DepPath = aMain

	bMain := filepath.Join(req.WorkRoot, labelCycleB)
	initGitRepoOnMain(t, bMain)
	writeGoModRequire(t, bMain, unwindCycleBModule, unwindCycleAModule+"@v0.0.0")
	writeFile(t, filepath.Join(bMain, "b.go"), "package cycleb\n")
	runGitIsolated(t, bMain, "add", "go.mod", "b.go")
	runGitIsolated(t, bMain, "commit", "-m", "add cycle-b")
	bMain = resolvePath(t, bMain)
	req.SecondRepo = bMain

	// Host checkout: neutral root that nests both externals (and may require both).
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule,
		unwindCycleAModule+"@v0.0.0",
		unwindCycleBModule+"@v0.0.0",
	)
	writeFile(t, filepath.Join(rootMain, "main.go"), "package main\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "main.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root host")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	wtDir := runWrkBin(t, req, rootMain, "--new")
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)

	aExt := filepath.Join(extDir, labelCycleA+"-"+branchNameMainDate())
	runGitIsolated(t, aMain, "worktree", "add", "-b", branchNameMainDate(), aExt)
	aExt = resolvePath(t, aExt)
	req.ExternalWtDir = aExt

	bExt := filepath.Join(extDir, labelCycleB+"-"+branchNameMainDate())
	runGitIsolated(t, bMain, "worktree", "add", "-b", branchNameMainDate(), bExt)
	bExt = resolvePath(t, bExt)
	req.DepsLinkedWtDir = bExt

	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	commitAllAllowEmpty(t, wtDir, "ignore external cycle members")

	// Dirty the cycle members so they enter pending (root host may stay clean).
	markDirty(t, aExt)
	markDirty(t, bExt)

	req.RepoDir = wtDir
}

func recordUnwindBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_unwind_baseline")
	mkdirAll(t, dir)
	if req.MainRepo != "" {
		writeFile(t, filepath.Join(dir, "main.sha"), revParseHEAD(t, req.MainRepo)+"\n")
	}
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			writeFile(t, filepath.Join(dir, "wt.sha"), revParseHEAD(t, req.WtDir)+"\n")
		}
	}
	if req.ExternalWtDir != "" {
		if _, err := os.Stat(req.ExternalWtDir); err == nil {
			writeFile(t, filepath.Join(dir, "ext.sha"), revParseHEAD(t, req.ExternalWtDir)+"\n")
		}
	}
	if req.DepsLinkedWtDir != "" {
		if _, err := os.Stat(req.DepsLinkedWtDir); err == nil {
			writeFile(t, filepath.Join(dir, "deps.sha"), revParseHEAD(t, req.DepsLinkedWtDir)+"\n")
		}
	}
}

func readBaselineSHA(t *testing.T, req *Request, name string) string {
	t.Helper()
	p := filepath.Join(req.WorkRoot, "_unwind_baseline", name)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertGitFileIsWorktreeLink(t *testing.T, wtDir string) {
	t.Helper()
	gitPath := filepath.Join(wtDir, ".git")
	info, err := os.Stat(gitPath)
	if os.IsNotExist(err) {
		t.Fatalf("%s should exist", gitPath)
	}
	if err != nil {
		t.Fatalf("stat %s: %v", gitPath, err)
	}
	if info.IsDir() {
		t.Fatalf("expected %s to be a regular file (linked worktree), got directory", gitPath)
	}
}

func assertUnwindZeroMutations(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo != "" {
		got := revParseHEAD(t, req.MainRepo)
		if want := readBaselineSHA(t, req, "main.sha"); got != want {
			t.Fatalf("main HEAD mutated: got %s want %s", got, want)
		}
	}
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			assertFileExists(t, req.WtDir)
			assertGitFileIsWorktreeLink(t, req.WtDir)
			got := revParseHEAD(t, req.WtDir)
			if want := readBaselineSHA(t, req, "wt.sha"); got != want {
				t.Fatalf("wt HEAD mutated: got %s want %s", got, want)
			}
		}
	}
	if req.ExternalWtDir != "" {
		if _, err := os.Stat(req.ExternalWtDir); err == nil {
			assertFileExists(t, req.ExternalWtDir)
			got := revParseHEAD(t, req.ExternalWtDir)
			if want := readBaselineSHA(t, req, "ext.sha"); got != want {
				t.Fatalf("external HEAD mutated: got %s want %s", got, want)
			}
		}
	}
	if req.DepsLinkedWtDir != "" {
		if _, err := os.Stat(req.DepsLinkedWtDir); err == nil {
			assertFileExists(t, req.DepsLinkedWtDir)
			got := revParseHEAD(t, req.DepsLinkedWtDir)
			if want := readBaselineSHA(t, req, "deps.sha"); got != want {
				t.Fatalf("deps external HEAD mutated: got %s want %s", got, want)
			}
		}
	}
	// single-main dirty leaves: DIRTY file must still exist (no apply clean)
	if req.WtDir == "" && req.MainRepo != "" {
		assertFileExists(t, filepath.Join(req.MainRepo, "DIRTY"))
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

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in %q", substr, s)
	}
}

// peelLine returns the locked dry-run peel line for a display path (not bare MainRepo basename).
func peelLine(displayPath string) string {
	return "would: peel " + displayPath
}

// applyBannerLine returns the locked apply peel banner for a display path.
func applyBannerLine(displayPath string) string {
	return "==== unwind: peel " + displayPath + " ===="
}

// assertPeelOrder checks free-first would: peel <display-path> lines appear in order.
// Also requires a dry-run/unwind banner somewhere in stdout.
func assertPeelOrder(t *testing.T, stdout string, displayPaths []string) {
	t.Helper()
	lower := strings.ToLower(stdout)
	if !strings.Contains(lower, "unwind") {
		t.Fatalf("stdout should mention unwind banner/plan, got %q", stdout)
	}
	if !strings.Contains(lower, "dry-run") && !strings.Contains(lower, "dry run") {
		// allow would: vocabulary alone if banner omits literal dry-run
		if !strings.Contains(stdout, "would: peel ") {
			t.Fatalf("stdout should look like dry-run peel plan, got %q", stdout)
		}
	}
	var prev int = -1
	for i, display := range displayPaths {
		line := peelLine(display)
		idx := strings.Index(stdout, line)
		if idx < 0 {
			t.Fatalf("missing peel line %q (step %d)\nstdout:\n%s", line, i+1, stdout)
		}
		if idx <= prev {
			t.Fatalf("peel order wrong at %q: idx=%d prev=%d\nstdout:\n%s", display, idx, prev, stdout)
		}
		prev = idx
	}
}

// assertNoBareBasenamePeelAlone fails if stdout uses bare MainRepo basename as the
// full peel path when a nested external display was expected (covers free-first RED).
func assertPeelUsesRelDisplay(t *testing.T, stdout, displayPath string) {
	t.Helper()
	if !strings.Contains(stdout, peelLine(displayPath)) {
		t.Fatalf("want peel display %q\nstdout:\n%s", peelLine(displayPath), stdout)
	}
	if displayPath == "." {
		// Primary at cwd must not be printed only as bare main-repo basename "root".
		if strings.Contains(stdout, peelLine(labelRoot)) {
			t.Fatalf("primary peel must be %q, not bare basename %q\nstdout:\n%s",
				peelLine("."), peelLine(labelRoot), stdout)
		}
		return
	}
	if strings.HasPrefix(displayPath, "external/") {
		// Nested external must include external/ prefix, not bare label alone as full path.
		// Bare "would: peel dot-pkgs" must not be the only form — require external/ fragment.
		if !strings.Contains(stdout, "would: peel external/") {
			t.Fatalf("nested peel must use external/ relative display, got:\n%s", stdout)
		}
	}
}

func assertNoSuccessfulPeelPlan(t *testing.T, stdout string) {
	t.Helper()
	// Cycle / hard reject must not emit a complete free-first success plan.
	if strings.Contains(stdout, "would: peel ") {
		// tolerate partial diagnostics only if exit non-zero already asserted;
		// still fail if multiple ordered peels look like a full plan.
		count := strings.Count(stdout, "would: peel ")
		if count >= 2 {
			t.Fatalf("cycle/error path must not print multi-step peel plan; stdout:\n%s", stdout)
		}
	}
}

func assertCycleError(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if !strings.Contains(combined, "cycle") {
		t.Fatalf("error must mention cycle; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	assertNoSuccessfulPeelPlan(t, resp.Stdout)
}

func assertMissingPinFlagsError(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "tag-next") && !strings.Contains(combined, "--tag-next") {
		t.Fatalf("error must mention --tag-next; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(lower, "push") && !strings.Contains(combined, "--push") {
		t.Fatalf("error must mention --push; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
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

func revParseRef(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func requireVersionInGoMod(t *testing.T, goModPath, modulePath string) string {
	t.Helper()
	content := readFile(t, goModPath)
	// Match "modulePath vX.Y.Z" on a require line (tolerant of tab/space).
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == modulePath {
			return fields[1]
		}
	}
	return ""
}

func goModHasReplace(t *testing.T, goModPath, modulePath string) bool {
	t.Helper()
	content := readFile(t, goModPath)
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") && strings.Contains(trim, modulePath) {
			return true
		}
		// multi-line replace block lines often start with the module path
		fields := strings.Fields(trim)
		if len(fields) >= 2 && fields[0] == modulePath && strings.Contains(content, "replace") {
			// only treat as replace when inside a replace context is hard; use substring
			if strings.Contains(trim, "=>") {
				return true
			}
		}
	}
	return strings.Contains(content, "replace "+modulePath) ||
		(strings.Contains(content, "replace (") && strings.Contains(content, modulePath+" =>"))
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func seedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)

	modContent := readFile(t, filepath.Join(srcDir, "go.mod"))
	writeFile(t, filepath.Join(vDir, version+".mod"), modContent)
	writeFile(t, filepath.Join(vDir, version+".info"),
		fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version))
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

// setupApplyLeafPinStack builds a 2-repo apply fixture:
//
//	leaf (dot-pkgs) main: tagged v0.0.1, bare origin, post-tag content for v0.0.2
//	leaf linked WT under root/external: commits ahead + DIRTY (pending + landable)
//	root main: requires leaf@v0.0.1, imports package; stays on main (pin target)
//
// RepoDir = root main. Flags set by leaf. Seeds modproxy for tidy after pin.
func setupApplyLeafPinStack(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf main: baseline tag + origin ---
	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "add dot-pkgs module")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain

	bare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, bare)
	runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	req.OriginBare = bare

	// --- root consumer on main (pin target; already main → no land for root) ---
	// Library package + blank-import keeps the require live without a main entrypoint.
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// Nest leaf linked WT under root/external (stack member).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	// Commit ahead on leaf WT (owned change → tag-next plans v0.0.2 after land).
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf feature for next tag")

	// Porcelain dirt so leaf enters dirty pending under v1 filter.
	// Implementer: commit dirt before --done land, or expand pending/ship pre-stage.
	markDirty(t, leafExt)

	// Ignore external/ on root so nesting does not dirty root porcelain.
	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "ignore external stack members")

	// Offline module proxy for pin+tidy (next tag tree = leaf WT tree).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, leafExt)
	// Old tag tree from leaf main baseline content snapshot.
	oldSeed := filepath.Join(req.WorkRoot, "seed-old-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = rootMain
	setPeelOrderDisplays(t, req, leafExt)
}

// setupApplyAlreadyMainRootBump: sole root main, tagged v0.0.1, owned change at HEAD,
// bare origin, porcelain DIRTY for pending. No stack edges → no land flags required.
// Args set by leaf: --unwind --tag-next --push.
// Library package root.go (no entrypoint) for owned-change tag-next lineage.
func setupApplyAlreadyMainRootBump(t *testing.T, req *Request) {
	t.Helper()

	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "root.go"), "package root\n\nfunc Version() string { return \"old\" }\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "root.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "add root module")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "")

	// Owned change after tag so tag-next plans v0.0.2.
	writeFile(t, filepath.Join(mainRepo, "root.go"), "package root\n\nfunc Version() string { return \"next\" }\n")
	runGitIsolated(t, mainRepo, "add", "root.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "owned change for tag-next")

	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo

	bare := setupBareOrigin(t, req.WorkRoot, "origin")
	attachOriginAndPushMain(t, mainRepo, bare)
	runGitIsolated(t, mainRepo, "push", "origin", unwindApplyOldTag)
	req.OriginBare = bare

	markDirty(t, mainRepo)
	req.PeelOrder = []string{"."}
	req.ExpectedPinVersion = unwindApplyNextTag
	req.OldRequireVersion = unwindApplyOldTag
}

func assertLocalTagAtMainHEAD(t *testing.T, mainRepo, tag string) {
	t.Helper()
	if !tagRefExists(t, mainRepo, tag) {
		t.Fatalf("local tag %s missing on %s", tag, mainRepo)
	}
	tagSHA := revParseRef(t, mainRepo, "refs/tags/"+tag)
	head := revParseHEAD(t, mainRepo)
	if tagSHA != head {
		t.Fatalf("tag %s at %s != main HEAD %s", tag, tagSHA, head)
	}
}

func assertOriginMainEqualsLocalMain(t *testing.T, mainRepo, originBare string) {
	t.Helper()
	mainSHA := revParseHEAD(t, mainRepo)
	originSHA := revParseRef(t, originBare, "refs/heads/main")
	if originSHA != mainSHA {
		t.Fatalf("origin/main %s != local main HEAD %s", originSHA, mainSHA)
	}
}

func assertConsumerPinned(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required for pin assert")
	}
	goMod := filepath.Join(req.MainRepo, "go.mod")
	got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("consumer require %s = %q, want %s\ngo.mod:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, readFile(t, goMod))
	}
	if goModHasReplace(t, goMod, req.LeafModulePath) {
		t.Fatalf("consumer go.mod still has replace for %s:\n%s",
			req.LeafModulePath, readFile(t, goMod))
	}
}

func assertLeafMainAdvancedAndTagged(t *testing.T, req *Request) {
	t.Helper()
	if req.SecondRepo == "" {
		t.Fatal("SecondRepo (leaf main) required")
	}
	// Leaf main should have received the feature commit from land.
	pkg := readFile(t, filepath.Join(req.SecondRepo, "pkg.go"))
	if !strings.Contains(pkg, unwindApplyNextTag) {
		t.Fatalf("leaf main pkg.go should include next-tag content; got:\n%s", pkg)
	}
	assertLocalTagAtMainHEAD(t, req.SecondRepo, req.ExpectedPinVersion)
	if req.OriginBare != "" {
		assertOriginMainEqualsLocalMain(t, req.SecondRepo, req.OriginBare)
		if !remoteTagExists(t, req.OriginBare, req.ExpectedPinVersion) {
			t.Fatalf("%s should exist on bare origin after --tag-next --push", req.ExpectedPinVersion)
		}
	}
}

func unwindEnsureHelpersUsed() {
	_ = setupSingleMainDirty
	_ = setupThreeRepoChain
	_ = dirtyAllThree
	_ = dirtyMidAndRoot
	_ = setupTwoCycleStack
	_ = setupApplyLeafPinStack
	_ = setupApplyAlreadyMainRootBump
	_ = recordUnwindBaseline
	_ = readBaselineSHA
	_ = assertUnwindZeroMutations
	_ = assertPeelOrder
	_ = assertPeelUsesRelDisplay
	_ = assertCycleError
	_ = assertMissingPinFlagsError
	_ = assertNoSuccessfulPeelPlan
	_ = assertExitZero
	_ = assertExitNonZero
	_ = assertErrIsNil
	_ = assertContains
	_ = assertNotContains
	_ = assertFileNotExists
	_ = assertConsumerPinned
	_ = assertLeafMainAdvancedAndTagged
	_ = assertLocalTagAtMainHEAD
	_ = assertOriginMainEqualsLocalMain
	_ = peelLine
	_ = applyBannerLine
	_ = peelDisplay
	_ = setPeelOrderDisplays
	_ = stageAll
	_ = leaveLine
	_ = assertContainsInOrder
	_ = assert.Output
}
```
