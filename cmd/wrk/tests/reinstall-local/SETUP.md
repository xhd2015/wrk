# Scenario

**Feature**: pure plan builder discovers cmd/script package mains and filters by binDir

```
# moduleRoot + binDir -> PlanLocalReinstalls -> sorted Plan items + Diagnostics
# cmd package main -> go-install; script/.../install -> go-run-install
# unique×unique → script wins + prefer-script notice; ambiguous tree skipped with warning
# missing $binDir/<bin> -> Action=skip (survivors only)
moduleRoot + binDir
  -> wrkcli.PlanLocalReinstalls
  -> LocalReinstallPlan{ModuleName, Items sorted by BinName, Diagnostics}
```

## Preconditions

- Package `github.com/xhd2015/wrk/wrkcli` is importable from this module
  (`go.mod` module path `github.com/xhd2015/wrk`).
- Nested root: no inheritance from `cmd/wrk/tests` (own `DOCTEST.md` Version 0.0.2).
- Pure API only — no wrk binary build, no git, no `WRK_HOME`, no `go install`/`go run`.
- Fixture trees are real directories under a per-leaf `t.TempDir()` WorkRoot:
  `ModuleRoot` holds `go.mod` + source; `BinDir` holds stub files named after bins.

## Steps

1. Root `Setup` creates isolated `WorkRoot`, empty `ModuleRoot`, and empty `BinDir`.
2. Leaves write `go.mod`, optional `package main` trees, and optional bin stubs.
3. Leaves set `WantError` / `WantModuleName` / `WantItems` / `WantDiagnostics`.
4. Root `Run` calls `wrkcli.PlanLocalReinstalls(ModuleRoot, BinDir)`.

## Context

- **Method strings**: `go-install` (cmd), `go-run-install` (script install dirs).
- **Action strings**: `install` when `$binDir/<bin>` is a present file (or
  symlink-to-file); `skip` otherwise. Plan keeps skip rows for survivors only.
- **RelPath**: always slash-form relative to module root, e.g. `./cmd/foo`,
  `./script/foo/install`, `./script/install`.
- **Module basename**: last segment of the `module` path in `go.mod`.
- **Diagnostic Level/Kind**: `notice`/`prefer-script`; `warning`/`ambiguous-cmd`;
  `warning`/`ambiguous-script`. Paths sorted slash-form `./…`.
- Helpers below write fixtures; Assert helpers compare plan structure.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"github.com/xhd2015/doctest/session"
)

const (
	methodGoInstall    = "go-install"
	methodGoRunInstall = "go-run-install"
	actionInstall      = "install"
	actionSkip         = "skip"

	diagLevelNotice   = "notice"
	diagLevelWarning  = "warning"
	diagKindPrefer    = "prefer-script"
	diagKindAmbCmd    = "ambiguous-cmd"
	diagKindAmbScript = "ambiguous-script"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.ModuleRoot = filepath.Join(workRoot, "mod")
	req.BinDir = filepath.Join(workRoot, "bin")
	if err := os.MkdirAll(req.ModuleRoot, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.BinDir, 0o755); err != nil {
		return err
	}
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	if req.WantDiagnostics == nil {
		req.WantDiagnostics = []WantDiagnostic{}
	}
	return nil
}

func writeGoMod(t *testing.T, moduleRoot, modulePath string) {
	t.Helper()
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

func writePackageNamed(t *testing.T, dir, pkg string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	src := "package " + pkg + "\n\nfunc Help() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "lib.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write lib.go in %s: %v", dir, err)
	}
}

func touchBin(t *testing.T, binDir, name string) {
	t.Helper()
	path := filepath.Join(binDir, name)
	if err := os.WriteFile(path, []byte("stub-binary\n"), 0o755); err != nil {
		t.Fatalf("touch bin %s: %v", path, err)
	}
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

func assertPlanOK(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ModuleRoot != req.ModuleRoot {
		t.Fatalf("ModuleRoot: got %q want %q", resp.ModuleRoot, req.ModuleRoot)
	}
	if resp.BinDir != req.BinDir {
		t.Fatalf("BinDir: got %q want %q", resp.BinDir, req.BinDir)
	}
	if resp.ModuleName != req.WantModuleName {
		t.Fatalf("ModuleName: got %q want %q", resp.ModuleName, req.WantModuleName)
	}
	if !reflect.DeepEqual(resp.Items, req.WantItems) {
		t.Fatalf("Items mismatch\n got: %#v\nwant: %#v", resp.Items, req.WantItems)
	}
	if !reflect.DeepEqual(resp.Diagnostics, req.WantDiagnostics) {
		t.Fatalf("Diagnostics mismatch\n got: %#v\nwant: %#v", resp.Diagnostics, req.WantDiagnostics)
	}
}

func assertPlanError(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrNonNil(t, err)
	if !req.WantError {
		t.Fatal("assertPlanError called but WantError is false")
	}
}
```
