# Scenario

**Feature**: wrk --pin-locals adds/normalizes relative replaces for already-required stack deps

```
# isolated WRK_HOME + git stack under WorkRoot
# inventory = primary + nested external/* (+ local-replace BFS)
stack modules + already-require/replace edges
  -> wrk --pin-locals [--dry-run]
  -> dry-run: would: pin-local …; no go.mod write; no tidy
  -> apply: pin-local …; relative replace; go mod tidy; summary stats
  -> tidy fail: warning: + continue; exit 0; stats show tidy failed
  -> exclusive / not-git: Error non-zero
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- Go toolchain and **git** on PATH.
- Per-leaf `t.TempDir()` WorkRoot; `WRK_HOME={WorkRoot}/.wrk`.
- L2: every leaf sets `req.InProcess = true` (wrkcli.Capture).
- Offline-friendly: apply leaves set `GOPROXY=off` + `GOSUMDB=off` via helpers
  when tidy must not hit the network (local replaces only).

## Steps

1. Root `Setup` creates isolated `WorkRoot` / `WrkHome`.
2. Leaves seed git + go.mod fixtures (multi-repo external, intra nested,
   abs rewrite, skip inventory, tidy-fail multi-consumer) and set `Args` /
   `RepoDir` / path fields.
3. Root `Run` invokes wrk via Capture when `InProcess` (default for leaves).

## Context

- **Module paths** used in fixtures:
  - `example.com/consumer` — primary consumer
  - `example.com/dep` — nested external stack dep
  - `example.com/root` / `example.com/root/tools` — intra-project
  - `example.com/other` — inventory-only non-dep
  - `example.com/missing` — require not owned by stack
  - `example.com/bad` — second consumer forced to fail tidy
- **Stdout pin line:** `pin-local <consumer> <- <dep> => <rel>`
- **Dry-run:** prefix `would: `
- **Summary:** `pin-locals: applied N, tidy ok M, tidy failed F`
- Nested externals under primary: gitignore `/external` so status treats them as
  independent nested repos (same as unwind/bring stack fixtures).

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/wrk/wrkcli"
)

const (
	wrkDate = "2026-06-30"

	modConsumer = "example.com/consumer"
	modDep      = "example.com/dep"
	modRoot     = "example.com/root"
	modTools    = "example.com/root/tools"
	modOther    = "example.com/other"
	modMissing  = "example.com/missing"
	modBad      = "example.com/bad"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	adoptDoctestContext(d)
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(workRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	// Default cwd until leaves reassign RepoDir.
	req.RepoDir = workRoot
	req.InProcess = true
	ensurePinLocalsHelpersUsed()
	return nil
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func resolvePath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %s: %v", p, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	if err := git_isolated.Init(path, "main"); err != nil {
		t.Fatalf("git init %s: %v", path, err)
	}
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func gitCommitAll(t *testing.T, repo, subject string) {
	t.Helper()
	runGitIsolated(t, repo, "add", "-A")
	runGitIsolated(t, repo, "commit", "-m", subject, "--allow-empty")
}

func writeGoMod(t *testing.T, dir, modulePath, body string) {
	t.Helper()
	mkdirAll(t, dir)
	content := "module " + modulePath + "\n\ngo 1.22\n"
	if body != "" {
		content += "\n" + body
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	}
	writeFile(t, filepath.Join(dir, "go.mod"), content)
}

func writeLibPkg(t *testing.T, dir, pkg, fn string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "pkg.go"),
		fmt.Sprintf("package %s\n\nfunc %s() string { return %q }\n", pkg, fn, fn))
}

// offlineEnv appends GOPROXY=off GOSUMDB=off so tidy never hits network.
func offlineEnv(req *Request) {
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY=off",
		"GOSUMDB=off",
	)
}

// setupMultiRepoExternalConsumer: primary git consumer requiring example.com/dep,
// nested independent git repo at external/dep (stack inventory member).
// No replace yet (apply/dry-run add path).
// RepoDir = primary root.
func setupMultiRepoExternalConsumer(t *testing.T, req *Request) {
	t.Helper()
	primary := filepath.Join(req.WorkRoot, "primary")
	initGitRepoOnMain(t, primary)
	writeGoMod(t, primary, modConsumer, "require "+modDep+" v0.0.1\n")
	writeLibPkg(t, primary, "consumer", "Hello")
	writeFile(t, filepath.Join(primary, ".gitignore"), "/external\n")
	gitCommitAll(t, primary, "consumer init")

	extRoot := filepath.Join(primary, "external")
	depDir := filepath.Join(extRoot, "dep")
	initGitRepoOnMain(t, depDir)
	writeGoMod(t, depDir, modDep, "")
	writeLibPkg(t, depDir, "dep", "Version")
	gitCommitAll(t, depDir, "dep init")

	// Re-commit primary after external/ appears so porcelain is clean
	// (external is gitignored; no consumer dirty required for pin-locals).
	gitCommitAll(t, primary, "ignore external")

	primary = resolvePath(t, primary)
	depDir = resolvePath(t, depDir)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.DepModDir = depDir
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	offlineEnv(req)
}

// setupAlreadyRelativePin: same as multi-repo but replace already points at
// relative ./external/dep (correct). Second plan should be already up to date.
func setupAlreadyRelativePin(t *testing.T, req *Request) {
	t.Helper()
	setupMultiRepoExternalConsumer(t, req)
	// Append correct relative replace.
	gm := readFile(t, req.ConsumerGoMod)
	if !strings.Contains(gm, "replace ") {
		gm = strings.TrimRight(gm, "\n") + "\n\nreplace " + modDep + " => ./external/dep\n"
		writeFile(t, req.ConsumerGoMod, gm)
		gitCommitAll(t, req.RepoDir, "already relative replace")
	}
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
}

// setupRewriteAbsolute: multi-repo with absolute replace NewPath to dep.
func setupRewriteAbsolute(t *testing.T, req *Request) {
	t.Helper()
	setupMultiRepoExternalConsumer(t, req)
	absDep := req.DepModDir
	gm := strings.TrimRight(req.BaselineGoMod, "\n") +
		"\n\nreplace " + modDep + " => " + absDep + "\n"
	writeFile(t, req.ConsumerGoMod, gm)
	gitCommitAll(t, req.RepoDir, "abs replace")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
}

// setupIntraProjectNested: single git repo, root requires nested tools module.
func setupIntraProjectNested(t *testing.T, req *Request) {
	t.Helper()
	primary := filepath.Join(req.WorkRoot, "primary")
	initGitRepoOnMain(t, primary)
	tools := filepath.Join(primary, "tools")
	writeGoMod(t, tools, modTools, "")
	writeLibPkg(t, tools, "tools", "Tool")
	writeGoMod(t, primary, modRoot, "require "+modTools+" v0.0.0\n")
	writeLibPkg(t, primary, "root", "Root")
	gitCommitAll(t, primary, "root+tools")

	primary = resolvePath(t, primary)
	tools = resolvePath(t, tools)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.ToolsModDir = tools
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	offlineEnv(req)
}

// setupSkipNotADependency: primary requires dep; other inventory module present
// under external/other but not required by consumer.
func setupSkipNotADependency(t *testing.T, req *Request) {
	t.Helper()
	setupMultiRepoExternalConsumer(t, req)

	otherDir := filepath.Join(req.RepoDir, "external", "other")
	initGitRepoOnMain(t, otherDir)
	writeGoMod(t, otherDir, modOther, "")
	writeLibPkg(t, otherDir, "other", "Other")
	gitCommitAll(t, otherDir, "other init")

	req.OtherModDir = resolvePath(t, otherDir)
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
}

// setupSkipNoMatchingOwner: primary requires example.com/missing only (not in stack).
// Optional inventory dep present but not required — pin should not invent edges.
func setupSkipNoMatchingOwner(t *testing.T, req *Request) {
	t.Helper()
	primary := filepath.Join(req.WorkRoot, "primary")
	initGitRepoOnMain(t, primary)
	// Require a path not owned by any stack module.
	writeGoMod(t, primary, modConsumer, "require "+modMissing+" v0.0.1\n")
	writeLibPkg(t, primary, "consumer", "Hello")
	writeFile(t, filepath.Join(primary, ".gitignore"), "/external\n")
	gitCommitAll(t, primary, "consumer missing-only")

	// Stack also has an unrelated dep module so inventory is non-empty.
	depDir := filepath.Join(primary, "external", "dep")
	initGitRepoOnMain(t, depDir)
	writeGoMod(t, depDir, modDep, "")
	writeLibPkg(t, depDir, "dep", "Version")
	gitCommitAll(t, depDir, "dep init")
	gitCommitAll(t, primary, "ignore external")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.DepModDir = resolvePath(t, depDir)
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	offlineEnv(req)
}

// setupTidyFailContinues: two consumer modules in the same primary checkout:
//   - root example.com/consumer requires dep (tidy ok after pin)
//   - nested bad/ example.com/bad requires dep + example.com/missing (tidy fails offline)
// Nested external dep at external/dep.
func setupTidyFailContinues(t *testing.T, req *Request) {
	t.Helper()
	primary := filepath.Join(req.WorkRoot, "primary")
	initGitRepoOnMain(t, primary)

	// Good consumer at root.
	writeGoMod(t, primary, modConsumer, "require "+modDep+" v0.0.1\n")
	writeLibPkg(t, primary, "consumer", "Hello")

	// Bad nested module: requires dep (will be pinned) + missing (no owner, no replace)
	// so go mod tidy fails with GOPROXY=off.
	badDir := filepath.Join(primary, "bad")
	writeGoMod(t, badDir, modBad,
		"require (\n\t"+modDep+" v0.0.1\n\t"+modMissing+" v0.0.1\n)\n")
	writeLibPkg(t, badDir, "bad", "Bad")

	writeFile(t, filepath.Join(primary, ".gitignore"), "/external\n")
	gitCommitAll(t, primary, "consumers init")

	depDir := filepath.Join(primary, "external", "dep")
	initGitRepoOnMain(t, depDir)
	writeGoMod(t, depDir, modDep, "")
	writeLibPkg(t, depDir, "dep", "Version")
	gitCommitAll(t, depDir, "dep init")
	gitCommitAll(t, primary, "ignore external")

	primary = resolvePath(t, primary)
	badDir = resolvePath(t, badDir)
	depDir = resolvePath(t, depDir)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BadModDir = badDir
	req.DepModDir = depDir
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	offlineEnv(req)
}

// runPinLocalsOnce invokes Capture once from Setup (idempotent second-run prelude).
func runPinLocalsOnce(t *testing.T, req *Request) *Response {
	t.Helper()
	args := append([]string(nil), req.Args...)
	if len(args) == 0 {
		args = []string{"--pin-locals"}
	}
	res := wrkcli.Capture(wrkcli.CaptureOpts{
		Args: args,
		Dir:  req.RepoDir,
		Env:  pinLocalsEnv(req),
	})
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}
}

// --- assert helpers ---

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

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in %q", substr, s)
	}
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

func assertMutualExclusion(t *testing.T, resp *Response) {
	t.Helper()
	assertExitNonZero(t, resp)
	se := resp.Stderr
	lower := strings.ToLower(se)
	if !strings.Contains(lower, "mutually exclusive") &&
		!strings.Contains(lower, "not valid") &&
		!strings.Contains(lower, "cannot") {
		t.Fatalf("stderr should indicate mutual exclusion, got %q", se)
	}
	if !strings.Contains(se, "--pin-locals") && !strings.Contains(se, "pin-locals") {
		// Prefer naming pin-locals; accept if product uses generic exclusive wording
		// that still includes the other flag — soft check via assert.Output.
		assert.Output(t, se, `<contains>
mutually exclusive
</contains>`)
	}
}

func assertGoModUnchanged(t *testing.T, req *Request) {
	t.Helper()
	if req.ConsumerGoMod == "" || req.BaselineGoMod == "" {
		t.Fatal("ConsumerGoMod and BaselineGoMod required")
	}
	got := readFile(t, req.ConsumerGoMod)
	if got != req.BaselineGoMod {
		t.Fatalf("go.mod mutated unexpectedly\n got:\n%s\nwant:\n%s", got, req.BaselineGoMod)
	}
}

// assertRelativeReplace checks go.mod has replace Old => relative New (./ or ../).
func assertRelativeReplace(t *testing.T, goModPath, oldPath string) {
	t.Helper()
	body := readFile(t, goModPath)
	// go mod edit -json would be ideal; parse loosely for replace lines.
	foundAbs := false
	foundRel := false
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "replace ") {
			continue
		}
		if !strings.Contains(trim, oldPath) {
			continue
		}
		// replace old => new  OR replace old => new (version forms not used for local)
		parts := strings.Split(trim, "=>")
		if len(parts) != 2 {
			continue
		}
		newPath := strings.TrimSpace(parts[1])
		// drop trailing version token if present
		fields := strings.Fields(newPath)
		if len(fields) > 0 {
			newPath = fields[0]
		}
		if strings.HasPrefix(newPath, "/") || filepath.IsAbs(newPath) {
			foundAbs = true
			continue
		}
		if strings.HasPrefix(newPath, "./") || strings.HasPrefix(newPath, "../") {
			foundRel = true
		}
	}
	if foundAbs && !foundRel {
		t.Fatalf("replace for %s is absolute only; want relative:\n%s", oldPath, body)
	}
	if !foundRel {
		t.Fatalf("expected relative replace for %s in:\n%s", oldPath, body)
	}
}

func assertNoReplaceFor(t *testing.T, goModPath, oldPath string) {
	t.Helper()
	body := readFile(t, goModPath)
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") && strings.Contains(trim, oldPath) {
			t.Fatalf("did not expect replace for %s:\n%s", oldPath, body)
		}
	}
}

func assertPinLocalLine(t *testing.T, stdout, consumer, dep string) {
	t.Helper()
	// pin-local <consumer> <- <dep> => <rel>
	needle := "pin-local " + consumer + " <- " + dep + " => "
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing pin-local line for %s <- %s; got:\n%s", consumer, dep, stdout)
	}
}

func assertWouldPinLocalLine(t *testing.T, stdout, consumer, dep string) {
	t.Helper()
	needle := "would: pin-local " + consumer + " <- " + dep + " => "
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing would: pin-local line for %s <- %s; got:\n%s", consumer, dep, stdout)
	}
}

func assertSummaryApplied(t *testing.T, stdout string, applied, tidyOK, tidyFailed int) {
	t.Helper()
	// Locked preferred form:
	// pin-locals: applied N, tidy ok M, tidy failed F
	want := fmt.Sprintf("pin-locals: applied %d, tidy ok %d, tidy failed %d", applied, tidyOK, tidyFailed)
	if !strings.Contains(stdout, want) {
		// Allow minor spacing variants but require the numbers.
		if !strings.Contains(stdout, fmt.Sprintf("applied %d", applied)) ||
			!strings.Contains(stdout, fmt.Sprintf("%d", tidyOK)) ||
			!strings.Contains(stdout, fmt.Sprintf("%d", tidyFailed)) {
			t.Fatalf("summary missing stats want %q; stdout:\n%s", want, stdout)
		}
		// If numbers present but phrasing differs, still fail to lock implementer.
		t.Fatalf("summary phrasing want %q; stdout:\n%s", want, stdout)
	}
}

func assertAlreadyUpToDate(t *testing.T, resp *Response) {
	t.Helper()
	assertExitZero(t, resp)
	all := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if !strings.Contains(all, "already") && !strings.Contains(all, "up to date") {
		t.Fatalf("expected already-up-to-date style message; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	// No successful apply pin-local lines (would: ok to be absent too).
	if strings.Contains(resp.Stdout, "pin-local ") &&
		!strings.Contains(resp.Stdout, "would: pin-local") {
		// bare pin-local apply lines should not appear when nothing to do
		// (summary may still say applied 0)
		for _, line := range strings.Split(resp.Stdout, "\n") {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "pin-local ") {
				t.Fatalf("did not expect apply pin-local line when already up to date: %q", trim)
			}
		}
	}
}

func assertWarningTidy(t *testing.T, stderr string) {
	t.Helper()
	if !strings.Contains(stderr, "warning:") {
		t.Fatalf("stderr must contain warning: for soft tidy fail, got %q", stderr)
	}
	lower := strings.ToLower(stderr)
	if !strings.Contains(lower, "go mod tidy") && !strings.Contains(lower, "mod tidy") {
		t.Fatalf("tidy warning must mention go mod tidy, got %q", stderr)
	}
}

func ensurePinLocalsHelpersUsed() {
	_ = setupMultiRepoExternalConsumer
	_ = setupAlreadyRelativePin
	_ = setupRewriteAbsolute
	_ = setupIntraProjectNested
	_ = setupSkipNotADependency
	_ = setupSkipNoMatchingOwner
	_ = setupTidyFailContinues
	_ = runPinLocalsOnce
	_ = assertRelativeReplace
	_ = assertNoReplaceFor
	_ = assertPinLocalLine
	_ = assertWouldPinLocalLine
	_ = assertSummaryApplied
	_ = assertAlreadyUpToDate
	_ = assertWarningTidy
	_ = assertGoModUnchanged
	_ = assertMutualExclusion
	_ = offlineEnv
	_ = writeGoMod
	_ = writeLibPkg
	_ = initGitRepoOnMain
	_ = gitCommitAll
}
```
