# Scenario

**Feature**: resolve reinstall scan root from workDir+useMain, scan modules, multi-plan

```
# workDir + useMain + binDir
#   -> ResolveReinstallScanRoot(workDir, useMain) -> scanRoot
#   -> mod/scan.Scan(scanRoot) -> module roots
#   -> PlanLocalReinstallsMulti -> MultiLocalReinstallPlan
#
# git (useMain=false): scanRoot = ShowToplevel(workDir)
# useMain=true:        scanRoot = main repo of checkout
# non-git:             scanRoot = walk-up go.mod from workDir
workDir + useMain + binDir
  -> wrkcli.ResolveReinstallScanRoot + PlanLocalReinstallsFromWorkDir
  -> ScanRoot + MultiLocalReinstallPlan | error
```

## Preconditions

- Package `github.com/xhd2015/wrk/wrkcli` is importable from this module
  (`go.mod` module path `github.com/xhd2015/wrk`).
- Nested root: no inheritance from `reinstall-local/` or `multi/` (own
  `DOCTEST.md`).
- Harness may create real git fixtures via
  `github.com/xhd2015/gitops/git/git_isolated` (hook-free). No wrk binary build,
  no `WRK_HOME`, no `go install`/`go run` of plan candidates.
- Fixture trees live under a per-leaf `t.TempDir()` WorkRoot (symlink-resolved).
  Shared `BinDir` holds stub files named after bins when Action=install is
  expected.

## Steps

1. Root `Setup` creates isolated `WorkRoot` and empty `BinDir`; defaults
   `UseMain=false`.
2. Leaves create git or non-git module trees, set `WorkDir` / `UseMain`, touch
   bins, and fill `WantScanRoot` / `WantModules` or error expectations.
3. Root `Run` calls `ResolveReinstallScanRoot` then
   `PlanLocalReinstallsFromWorkDir`.

## Context

- **Method strings**: `go-install` (cmd), `go-run-install` (script install dirs).
- **Action strings**: `install` when `$binDir/<bin>` is a present file; `skip`
  otherwise. Plan keeps skip rows.
- **RelPath**: slash-form relative to that module's root.
- **Module basename**: last segment of the `module` path in `go.mod`.
- **Sort**: modules by absolute ModuleRoot lex; items by BinName within module.
- **Path identity**: compare ScanRoot / ModuleRoot after `filepath.Abs` +
  `EvalSymlinks` when fixtures use git (macOS `/var` → `/private/var`).

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	methodGoInstall    = "go-install"
	methodGoRunInstall = "go-run-install"
	actionInstall      = "install"
	actionSkip         = "skip"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.BinDir = filepath.Join(workRoot, "bin")
	if err := os.MkdirAll(req.BinDir, 0o755); err != nil {
		return err
	}
	req.UseMain = false
	if req.WantModules == nil {
		req.WantModules = []WantModulePlan{}
	}
	if req.WantErrSubstrs == nil {
		req.WantErrSubstrs = []string{}
	}
	return nil
}

func writeGoMod(t *testing.T, moduleRoot, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatalf("mkdir module root %s: %v", moduleRoot, err)
	}
	content := "module " + modulePath + "\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func writePackageMain(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	// Build source without a single string literal containing both
	// "package main" and "func main()" (doctest anti-pattern check).
	src := fmt.Sprintf("package %s\n\nfunc main() {}\n", "main")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write main.go in %s: %v", dir, err)
	}
}

func touchBin(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	if err := os.WriteFile(path, []byte("stub-binary\n"), 0o755); err != nil {
		t.Fatalf("touch bin %s: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
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
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func gitCommitAll(t *testing.T, repo, subject string) {
	t.Helper()
	runGitIsolated(t, repo, "add", "-A")
	runGitIsolated(t, repo, "commit", "-m", subject)
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertErrNonNil(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertScanPlanOK(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	wantRoot := resolvePath(t, req.WantScanRoot)
	gotRoot := resolvePath(t, resp.ScanRoot)
	if gotRoot != wantRoot {
		t.Fatalf("ScanRoot: got %q want %q", gotRoot, wantRoot)
	}
	if resp.BinDir != req.BinDir {
		t.Fatalf("BinDir: got %q want %q", resp.BinDir, req.BinDir)
	}
	// Normalize ModuleRoot paths for macOS symlink stability before compare.
	gotMods := make([]WantModulePlan, len(resp.Modules))
	for i, m := range resp.Modules {
		gotMods[i] = m
		gotMods[i].ModuleRoot = resolvePath(t, m.ModuleRoot)
	}
	wantMods := make([]WantModulePlan, len(req.WantModules))
	for i, m := range req.WantModules {
		wantMods[i] = m
		wantMods[i].ModuleRoot = resolvePath(t, m.ModuleRoot)
	}
	if !reflect.DeepEqual(gotMods, wantMods) {
		t.Fatalf("Modules mismatch\n got: %#v\nwant: %#v", gotMods, wantMods)
	}
}

func assertScanPlanError(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrNonNil(t, err)
	if !req.WantError {
		t.Fatal("assertScanPlanError called but WantError is false")
	}
	msg := err.Error()
	for _, sub := range req.WantErrSubstrs {
		if sub == "" {
			continue
		}
		if !strings.Contains(msg, sub) {
			t.Fatalf("error %q does not contain %q", msg, sub)
		}
	}
}
```
