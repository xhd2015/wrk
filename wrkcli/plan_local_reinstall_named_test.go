package wrkcli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterNamedReinstallPlan_ForceInstallAndSelect(t *testing.T) {
	mod := t.TempDir()
	writeTestGoMod(t, mod, "example.com/named")
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "alpha"))
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "beta"))
	binDir := t.TempDir()
	// Only alpha is present in binDir → full plan would skip beta.
	if err := os.WriteFile(filepath.Join(binDir, "alpha"), []byte("stub\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	full, err := PlanLocalReinstallsMulti([]string{mod}, binDir)
	if err != nil {
		t.Fatal(err)
	}

	named, err := filterNamedReinstallPlan(full, []string{"beta", "alpha", "beta"})
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
	// Request order after dedupe: beta, alpha.
	if items[0].BinName != "beta" || items[0].Action != ActionInstall {
		t.Fatalf("first item=%+v want beta install", items[0])
	}
	if items[1].BinName != "alpha" || items[1].Action != ActionInstall {
		t.Fatalf("second item=%+v want alpha install", items[1])
	}
}

func TestFilterNamedReinstallPlan_Unknown(t *testing.T) {
	mod := t.TempDir()
	writeTestGoMod(t, mod, "example.com/named")
	writeTestPackageMain(t, filepath.Join(mod, "cmd", "only"))
	binDir := t.TempDir()

	full, err := PlanLocalReinstallsMulti([]string{mod}, binDir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = filterNamedReinstallPlan(full, []string{"missing"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no install candidate") || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("error=%v", err)
	}
}

func TestFilterNamedReinstallPlan_CrossModuleCollision(t *testing.T) {
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
	// Both are skip in full plan (no bin), so Multi did not collide yet.
	_, err = filterNamedReinstallPlan(full, []string{"same"})
	if err == nil {
		t.Fatal("expected collision")
	}
	if !strings.Contains(err.Error(), "multiple modules") || !strings.Contains(err.Error(), "same") {
		t.Fatalf("error=%v", err)
	}
}

func writeTestGoMod(t *testing.T, dir, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "module " + modulePath + "\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeTestPackageMain(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}
