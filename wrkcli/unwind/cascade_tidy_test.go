package unwind

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStackMainDirPrefersMainRepo(t *testing.T) {
	t.Parallel()
	by := map[string]StackMember{
		"leaf": {Path: "/tmp/leaf-wt", MainRepo: "/tmp/leaf-main", Label: "leaf"},
		"only": {Path: "/tmp/only", Label: "only"},
	}
	if got := stackMainDir(by, "leaf"); got != "/tmp/leaf-main" {
		t.Fatalf("stackMainDir leaf=%q want main", got)
	}
	if got := stackMainDir(by, "only"); got != "/tmp/only" {
		t.Fatalf("stackMainDir only=%q want path", got)
	}
	if got := stackMainDir(by, "missing"); got != "" {
		t.Fatalf("stackMainDir missing=%q", got)
	}
}

func TestGoModTidyForCascadePinSkipsVendor(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := goModTidyForCascadePin(dir, goModSumSnap{}, false, "github.com/xhd2015/lib", dir); err != nil {
		t.Fatalf("vendor skip: %v", err)
	}
}

func TestGoModTidyForCascadePinUsesLocalGit(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	scene := t.TempDir()
	modCache := filepath.Join(scene, "gomodcache")
	t.Cleanup(func() {
		_ = filepath.Walk(modCache, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			_ = os.Chmod(path, 0o755)
			return nil
		})
	})
	lib := filepath.Join(scene, "lib")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "go.mod"), []byte("module github.com/xhd2015/wrk-unwind-tidy\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "hello.go"), []byte("package unwindingtidy\n\nfunc Hello() string { return \"hi\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = lib
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-b", "main")
	git("config", "user.name", "expt")
	git("config", "user.email", "expt@example.com")
	git("add", ".")
	git("commit", "-m", "init")
	git("tag", "v0.0.1")

	consume := filepath.Join(scene, "consume")
	if err := os.MkdirAll(consume, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consume, "go.mod"), []byte(""+
		"module consume\n\n"+
		"go 1.22\n\n"+
		"require github.com/xhd2015/wrk-unwind-tidy v0.0.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consume, "main.go"), []byte(""+
		"package main\n\n"+
		"import (\n"+
		"\t\"fmt\"\n"+
		"\tunwindingtidy \"github.com/xhd2015/wrk-unwind-tidy\"\n"+
		")\n\n"+
		"func main() { fmt.Print(unwindingtidy.Hello()) }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOMODCACHE", modCache)
	t.Setenv("GOCACHE", filepath.Join(scene, "gocache"))
	t.Setenv("GOPATH", filepath.Join(scene, "gopath"))
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("https_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("http_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	t.Setenv("GOPROXY", "off")

	if err := goModTidyForCascadePin(consume, goModSumSnap{}, false, "github.com/xhd2015/wrk-unwind-tidy", lib); err != nil {
		t.Fatalf("tidy: %v", err)
	}
	sum, err := os.ReadFile(filepath.Join(consume, "go.sum"))
	if err != nil {
		t.Fatalf("go.sum: %v", err)
	}
	if !strings.Contains(string(sum), "github.com/xhd2015/wrk-unwind-tidy v0.0.1") {
		t.Fatalf("go.sum missing module:\n%s", sum)
	}
}
