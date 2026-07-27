# Scenario

**Feature**: pure multi-module plan builder over an explicit module-root list

```
# moduleRoots[] + binDir -> PlanLocalReinstallsMulti
# per-module discovery (cmd/script rules) + install×install cross-module hard error
moduleRoots + binDir
  -> wrkcli.PlanLocalReinstallsMulti
  -> MultiLocalReinstallPlan{BinDir, Modules sorted by ModuleRoot}
     | error (install×install collision or invalid module)
```

## Preconditions

- Package `github.com/xhd2015/wrk/wrkcli` is importable from this module
  (`go.mod` module path `github.com/xhd2015/wrk`).
- Nested root: no inheritance from `reinstall-local/` parent (own `DOCTEST.md`).
- Pure API only — no wrk binary build, no git, no `WRK_HOME`, no `go install`/`go run`.
- Fixture trees are real directories under a per-leaf `t.TempDir()` WorkRoot:
  each module root holds `go.mod` + source; shared `BinDir` holds stub files
  named after bins.

## Steps

1. Root `Setup` creates isolated `WorkRoot` and empty `BinDir`.
2. Leaves create one or more module directories under `WorkRoot`, write `go.mod`
   and optional `package main` trees, touch bin stubs, set `ModuleRoots`.
3. Leaves set `WantError` / `WantErrSubstrs` / `WantModules`.
4. Root `Run` calls `wrkcli.PlanLocalReinstallsMulti(ModuleRoots, BinDir)`.

## Context

- **Method strings**: `go-install` (cmd), `go-run-install` (script install dirs).
- **Action strings**: `install` when `$binDir/<bin>` is a present file (or
  symlink-to-file); `skip` otherwise. Plan keeps skip rows.
- **RelPath**: always slash-form relative to that module's root.
- **Module basename**: last segment of the `module` path in `go.mod`.
- **Sort**: modules by absolute ModuleRoot lex; items by BinName within module.
- Helpers below write fixtures; Assert helpers compare multi-plan structure.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
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
	if req.ModuleRoots == nil {
		req.ModuleRoots = []string{}
	}
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

func assertMultiPlanOK(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.BinDir != req.BinDir {
		t.Fatalf("BinDir: got %q want %q", resp.BinDir, req.BinDir)
	}
	if !reflect.DeepEqual(resp.Modules, req.WantModules) {
		t.Fatalf("Modules mismatch\n got: %#v\nwant: %#v", resp.Modules, req.WantModules)
	}
}

func assertMultiPlanError(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrNonNil(t, err)
	if !req.WantError {
		t.Fatal("assertMultiPlanError called but WantError is false")
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
