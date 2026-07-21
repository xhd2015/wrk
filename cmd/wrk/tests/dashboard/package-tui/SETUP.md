# Scenario

**Feature**: interactive dashboard TUI extracted to package `github.com/xhd2015/wrk/wrkcli/tui`

```
# package layout (extraction boundary)
wrk CLI / wrkcli
  -> interactive Bubble Tea lives in wrkcli/tui
  -> public entry: tui.RunDashboard(opts RunDashboardOpts) error
  -> public types: tui.Recipe, tui.RunDashboardOpts
  -> tui must not import parent package wrkcli (callbacks/opts inject deps)

# classic TDD
  -> until implementer lands wrkcli/tui, package leaves RED
  -> existing dashboard hermetic/snapshot/tty/docs leaves stay the behavior seal
```

## Preconditions

- Go toolchain on PATH (same as root harness).
- Module root is an ancestor of `d.DOCTEST_ROOT` (has `go.mod` for `github.com/xhd2015/wrk`).
- No git repo required; package leaves do not exercise live TUI I/O.
- Shared `Request` / `Run` from `cmd/wrk/tests/DOCTEST.md` still invoke the `wrk` binary once per leaf (cheap `-h`) so the harness stays uniform; **assertions are package-surface checks**, not CLI UX.

## Steps

- Grouping installs helpers: resolve module root, import path constant, `go list` / `go doc` wrappers.
- Leaves: **importable** (package path exists) → **run-dashboard-exported** (public API) → **no-import-cycle** (dependency direction).
- Significance order: existence → exports → cycle (largest package-boundary impact first).

## Context

- Expected import path after implementer: `github.com/xhd2015/wrk/wrkcli/tui`.
- Expected public surface (field names may be exported Go style; assert **symbols**, not private helpers):

  - `func RunDashboard(opts RunDashboardOpts) error`
  - `type Recipe struct { … }`
  - `type RunDashboardOpts struct { … }` (workdir, status, injectables)

- Cycle rule: `wrkcli/tui` may import siblings like `wrkcli/teapre` but **must not** import parent `github.com/xhd2015/wrk/wrkcli`.
- Behavior seal remains under `snapshot/`, `interactive/`, `tty/`, `docs/`, etc. — do not weaken those leaves.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const wrkcliTuiImportPath = "github.com/xhd2015/wrk/wrkcli/tui"

// wrkcliParentImportPath is the parent package tui must not import (cycle).
const wrkcliParentImportPath = "github.com/xhd2015/wrk/wrkcli"

func packageTuiModuleRoot(t *testing.T, d *session.Doctest) string {
	t.Helper()
	dir := d.DOCTEST_ROOT
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
		}
		dir = parent
	}
}

// goListInModule runs `go list` with extra args in the wrk module root.
// Returns combined stdout+stderr and the process error (non-nil when exit != 0).
func goListInModule(t *testing.T, d *session.Doctest, args ...string) (string, error) {
	t.Helper()
	modRoot := packageTuiModuleRoot(t, d)
	cmdArgs := append([]string{"list"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = modRoot
	// Keep env mostly ambient; pin module mode for hermetic list.
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// goDocInModule runs `go doc` for a package or package.Symbol in the module.
func goDocInModule(t *testing.T, d *session.Doctest, target string) (string, error) {
	t.Helper()
	modRoot := packageTuiModuleRoot(t, d)
	cmd := exec.Command("go", "doc", target)
	cmd.Dir = modRoot
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func assertPackageListed(t *testing.T, d *session.Doctest, importPath string) {
	t.Helper()
	out, err := goListInModule(t, d, importPath)
	if err != nil {
		t.Fatalf("package %s must be listable in module (go list): %v\n%s",
			importPath, err, out)
	}
	got := strings.TrimSpace(out)
	if got != importPath {
		t.Fatalf("go list %s: want exact import path %q, got %q", importPath, importPath, got)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Package-boundary leaves: no git fixture; cheap wrk -h satisfies root Run.
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-h"}
	req.TargetDir = ""
	req.TaskDesc = ""
	req.ExtraEnv = nil
	// Touch helpers so unused-import / deadcode checkers stay quiet if a leaf only uses a subset.
	_ = filepath.Separator
	return nil
}

```
