# Scenario

**Feature**: `wrk --done` UX — phase banners only when cascade targets ≥ 1; structured cascade `Error:` with path context

```
# zero cascade targets: no phase headers; still print real result
linked wt (no nested cascade)
  -> wrk --done [--dry-run]
  -> (no ==> cascade, no ==> own)
  -> primary MergeBack plan or apply body only

# with cascade targets (≥1): frame cascade then own
linked wt + nested linked cascade targets
  -> wrk --done [--dry-run]
  -> ==> cascade
  -> cascade items / would: cascade merge-back <path>
  -> ==> own
  -> primary MergeBack plan or apply

# cascade MergeBack failure (e.g. diverged rebase conflict)
  -> ==> cascade (present; non-empty cascade)
  -> non-zero exit mid-cascade
  -> stderr framed with Error: + external path (not bare "rebase conflict:" alone)
  -> ==> own may be absent (own phase never reached)
```

## Preconditions

- Inherits monotree **root** helpers only (`setupWrkWorktreeFromMain`, `commitAheadOnWorktree`,
  `runWrkFrom` / `runWrkWithArgs`, `compositionResolvePath`, git helpers).
  Do **not** call `done-pipeline/`-scoped helpers (sibling tree; not in inheritance chain).
- Classic TDD: production currently always prints both phase headers even with zero cascade
  targets — **zero-cascade** leaves expect **RED** until implementer skips headers when
  cascade count is 0 (`runDone` / phase-print path). Structured cascade `Error:` and
  with-cascade headers may already be GREEN.
- Prefer structure/text asserts (substring / order). Do **not** require ANSI color sequences.

## Steps

- Grouping only: leaves build fixtures and set `req.Args` / `req.RepoDir`.

```go
import (
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

// assertDonePhaseHeaders checks stderr∪stdout for phase banners (stream-flexible).
// Use when cascade targets ≥ 1: both ==> cascade and ==> own; cascade before own.
func assertDonePhaseHeaders(t *testing.T, resp *Response) {
	t.Helper()
	combined := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(combined, "==> cascade") {
		t.Fatalf("missing phase header %q in stdout/stderr\nstdout:\n%s\nstderr:\n%s",
			"==> cascade", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(combined, "==> own") {
		t.Fatalf("missing phase header %q in stdout/stderr\nstdout:\n%s\nstderr:\n%s",
			"==> own", resp.Stdout, resp.Stderr)
	}
	idxCascade := strings.Index(combined, "==> cascade")
	idxOwn := strings.Index(combined, "==> own")
	if idxCascade < 0 || idxOwn < 0 || idxCascade > idxOwn {
		t.Fatalf("expected ==> cascade before ==> own; cascade@%d own@%d\ncombined:\n%s",
			idxCascade, idxOwn, combined)
	}
}

// assertNoDonePhaseHeaders fails if either phase banner is printed (zero-cascade path).
func assertNoDonePhaseHeaders(t *testing.T, resp *Response) {
	t.Helper()
	combined := resp.Stdout + "\n" + resp.Stderr
	if strings.Contains(combined, "==> cascade") {
		t.Fatalf("zero-cascade must not print %q; stdout:\n%s\nstderr:\n%s",
			"==> cascade", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(combined, "==> own") {
		t.Fatalf("zero-cascade must not print %q; stdout:\n%s\nstderr:\n%s",
			"==> own", resp.Stdout, resp.Stderr)
	}
}

// assertNoANSIEscape fails if output contains CSI color sequences (structure-only UX pin).
func assertNoANSIEscape(t *testing.T, s, label string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[") {
		t.Fatalf("%s must not require ANSI for structure asserts (found escape); got:\n%s", label, s)
	}
}

// setupDoneOutputLocalAhead: simple linked wt ahead of main (no nested cascade targets).
func setupDoneOutputLocalAhead(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// assertPrimaryDoneDryRunPlanned: MergeBack DryRun shapes for ahead+remove (path-agnostic).
func assertPrimaryDoneDryRunPlanned(t *testing.T, stdout, wtBranch string) {
	t.Helper()
	if strings.Contains(stdout, "Proceed?") {
		t.Fatalf("dry-run must not prompt; stdout=%q", stdout)
	}
	assertContains(t, stdout, "merge --ff-only "+wtBranch)
	assertContains(t, stdout, "worktree remove")
	assertContains(t, stdout, "branch -D "+wtBranch)
}

// assertNoConfirmPromptNoiseUX: stderr/stdout free of confirm / non-tty prompt errors.
func assertNoConfirmPromptNoiseUX(t *testing.T, resp *Response) {
	t.Helper()
	assertNotContains(t, resp.Stdout, "Proceed?")
	assertNotContains(t, resp.Stderr, "Proceed?")
	assertNotContains(t, resp.Stderr, "stdin is not a terminal")
	assertNotContains(t, resp.Stderr, "cannot prompt")
	assertNotContains(t, resp.Stdout, "cannot prompt")
}

// setupDivergedExternalForCascadeFail: consumer wt + external dep wt with diverged
// histories on dep.go so cascade MergeBack hits rebase conflict (hard fail).
func setupDivergedExternalForCascadeFail(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	const mod = "example.com/dep-ux-fail"
	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, mainRepo, "edit", "-require="+mod+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+mod+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n// base\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--bring", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	// Commit consumer porcelain after --bring so D2 dirty preflight does not
	// block before the cascade MergeBack conflict under test.
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "commit dep replace and ignore", "--allow-empty")

	// Diverged: main and external each edit the same lines of dep.go.
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n// main-side change\n")
	runGitIsolated(t, depRepo, "add", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "main-side conflicting change")

	writeFile(t, filepath.Join(externalPath, "dep.go"), "package dep\n// external-side change\n")
	runGitIsolated(t, externalPath, "add", "dep.go")
	runGitIsolated(t, externalPath, "commit", "-m", "external-side conflicting change")
}

var (
	_ = assertDonePhaseHeaders
	_ = assertNoDonePhaseHeaders
	_ = assertNoANSIEscape
	_ = setupDoneOutputLocalAhead
	_ = assertPrimaryDoneDryRunPlanned
	_ = assertNoConfirmPromptNoiseUX
	_ = setupDivergedExternalForCascadeFail
)
```
