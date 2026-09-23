package wrkcli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildInstallTestPlan writes a module with cmd/alpha + cmd/beta, marks only
// alpha present in binDir, and returns the full (binDir-gated) multi plan.
func buildInstallTestPlan(t *testing.T, mod, binDir string) *MultiLocalReinstallPlan {
	t.Helper()
	writeTestGoMod(t, mod, "example.com/install")
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "alpha"))
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "beta"))
	if err := os.WriteFile(filepath.Join(binDir, "alpha"), []byte("stub\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	full, err := PlanLocalReinstallsMulti([]string{mod}, binDir)
	if err != nil {
		t.Fatal(err)
	}
	return full
}

func TestLookupNamedPlan_InstallModeForcesAbsentBin(t *testing.T) {
	mod := t.TempDir()
	binDir := t.TempDir()
	full := buildInstallTestPlan(t, mod, binDir)

	named, err := lookupNamedPlan(full, []string{"beta", "alpha", "beta"}, ModeInstall)
	if err != nil {
		t.Fatal(err)
	}
	if len(named.Modules) != 1 {
		t.Fatalf("modules=%d want 1", len(named.Modules))
	}
	items := named.Modules[0].Items
	if len(items) != 2 {
		t.Fatalf("items=%v want alpha+beta", items)
	}
	// Request order after dedupe: beta, alpha — both forced to install even
	// though only alpha is present in binDir.
	if items[0].BinName != "beta" || items[0].Action != ActionInstall {
		t.Fatalf("first item=%+v want beta install", items[0])
	}
	if items[1].BinName != "alpha" || items[1].Action != ActionInstall {
		t.Fatalf("second item=%+v want alpha install", items[1])
	}
}

func TestLookupNamedPlan_InstallModeUnknown(t *testing.T) {
	mod := t.TempDir()
	binDir := t.TempDir()
	full := buildInstallTestPlan(t, mod, binDir)

	_, err := lookupNamedPlan(full, []string{"missing"}, ModeInstall)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "wrk: --install:") {
		t.Fatalf("error should carry the --install prefix: %v", err)
	}
	if !strings.Contains(msg, "no install candidate") || !strings.Contains(msg, "missing") {
		t.Fatalf("error=%v", err)
	}
}

func TestLookupNamedPlan_InstallModeCrossModuleCollision(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	writeTestGoMod(t, a, "example.com/a")
	writeTestGoMod(t, b, "example.com/b")
	writeTestPackageMain(t, filepath.Join(a, "cmd", "same"))
	writeTestPackageMain(t, filepath.Join(b, "cmd", "same"))
	binDir := t.TempDir()

	full, err := PlanLocalReinstallsMulti([]string{a, b}, binDir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = lookupNamedPlan(full, []string{"same"}, ModeInstall)
	if err == nil {
		t.Fatal("expected collision")
	}
	msg := err.Error()
	if !strings.Contains(msg, "wrk: --install:") {
		t.Fatalf("error should carry the --install prefix: %v", err)
	}
	if !strings.Contains(msg, "multiple modules") || !strings.Contains(msg, "same") {
		t.Fatalf("error=%v", err)
	}
}

// TestRunInstallExTo_DryRunVocabulary locks the --install stdout vocabulary:
// the shared engine prints "would: go install <path>" plus "would: install N
// binaries", and never a skip: line (named install is always forced).
func TestRunInstallExTo_DryRunVocabulary(t *testing.T) {
	mod := t.TempDir()
	binDir := t.TempDir()
	writeTestGoMod(t, mod, "example.com/install-dry")
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "tool"))
	t.Setenv("GOBIN", binDir)

	var out, errW bytes.Buffer
	st, err := runInstallExTo(mod, true, false, false, false, []string{"tool"}, &out, &errW)
	if err != nil {
		t.Fatal(err)
	}
	if st != (ReinstallExecStats{}) {
		t.Fatalf("dry-run stats=%+v want zero", st)
	}
	want := "would: go install ./cmd/tool\nwould: install 1 binaries\n"
	if out.String() != want {
		t.Fatalf("stdout\n got: %q\nwant: %q", out.String(), want)
	}
	if strings.Contains(out.String(), "skip:") {
		t.Fatalf("named install must not skip: %q", out.String())
	}
	if errW.String() != "" {
		t.Fatalf("stderr should be empty, got %q", errW.String())
	}
}

// TestRunInstallExTo_RequiresName guards the defensive check behind the
// parser-level WithMinimum(1).
func TestRunInstallExTo_RequiresName(t *testing.T) {
	mod := t.TempDir()
	writeTestGoMod(t, mod, "example.com/install-empty")
	t.Setenv("GOBIN", t.TempDir())

	_, err := runInstallExTo(mod, true, false, false, false, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "requires at least one name") {
		t.Fatalf("error=%v", err)
	}
}
