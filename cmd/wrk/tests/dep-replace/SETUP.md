# Scenario

**Feature**: wrk --dep-replace stack fan-out + CLI tree (absolute + versioned tidy)

```
# isolated WRK_HOME; nearest not-git fallback or CollectStackInventory
consumer go.mod + dep module dir(s)
  -> wrk --dep-replace <dir>… [--dry-run]
  -> dry-run: ==== dep-replace (dry-run) ====; would: replace + would: go mod tidy; no write
  -> apply: ==== dep-replace ====; checkout → module → replace → go mod tidy ok
  -> git stack: gate require|existing replace; self skipped
  -> not-git nearest: D7 write even with no require
  -> exclusive / empty / missing / not-module / zero gated: wrk: error; no banner
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- Go toolchain on PATH (uses `go mod edit` + versioned tidy via `withgo`).
- Per-leaf `t.TempDir()` WorkRoot; `WRK_HOME={WorkRoot}/.wrk`.
- L2: every leaf sets `req.InProcess = true` (wrkcli.Capture). `Run` passes `CaptureOpts.WithGo`.
- Root Setup seeds `$InstallDir/go1.22.12/bin/go` host-go wrapper (`go 1.22` pin).
- Existing nearest leaves stay **non-git** (D7 walk-up fallback). Stack leaves
  init a primary git repo plus an independent `external/kool` (or
  `external/dep`) git repo and a **local filesystem replace** so
  `CollectStackInventory` BFS includes it.
- Offline-friendly: local `replace` targets; `WithGo.Download=false`.

## Steps

1. Root `Setup` creates isolated `WorkRoot` / `WrkHome` and seeds the `go1.22.12` wrapper.
2. Leaves seed consumer + dep go.mod fixtures and set `Args` / `RepoDir` / paths.
3. Root `Run` invokes wrk via Capture with `WithGo` when `InProcess`.

## Context

- **Module paths** used in fixtures:
  - `example.com/consumer` — nearest consumer (not-git apply leaves)
  - `example.com/app` — stack primary
  - `example.com/kool` — stack other-checkout at `external/kool`
  - `example.com/dep` — primary dep
  - `example.com/dep2` — second dep (multi-dir)
- **Stdout apply:** `==== dep-replace ====`; `dep` headers; checkout →
  module → `replace` + `go mod tidy ok` or `skip tidy  (vendor/)`;
  `dep-replace: replaced in N modules in C checkouts`
- **Dry-run:** `==== dep-replace (dry-run) ====`; `would: replace` +
  `would: go mod tidy` / `would: skip tidy  (vendor/)`;
  `would replace in N modules in C checkouts`
- **Not-git consumer resolution:** walk up from `RepoDir` to nearest go.mod.
- **Git consumer set:** `CollectStackInventory(cwd)`.
- Absolute NewPath must match resolved dep dir (EvalSymlinks-normalized when needed).

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	wrkDate = "2026-06-30"

	modConsumer = "example.com/consumer"
	modDep      = "example.com/dep"
	modDep2     = "example.com/dep2"

	modApp       = "example.com/app"
	modKool      = "example.com/kool"
	checkoutKool = "external/kool"
	checkoutDep  = "external/dep"

	pinGo122 = "go1.22.12"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	adoptDoctestContext(d)
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	registerRetryWorkRootCleanup(t, workRoot)
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(workRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	req.RepoDir = workRoot
	req.InProcess = true
	req.InstallDir = filepath.Join(workRoot, "installed")
	goPath := filepath.Join(workRoot, "go")
	modCache := filepath.Join(goPath, "pkg", "mod")
	goCache := filepath.Join(workRoot, "gocache")
	req.ExtraEnv = append(req.ExtraEnv,
		"HOME="+workRoot,
		"GOTELEMETRY=off",
		"GOPATH="+goPath,
		"GOMODCACHE="+modCache,
		"GOCACHE="+goCache,
	)
	// go may fill caches with read-only files; unlock before t.TempDir cleanup.
	t.Cleanup(func() {
		for _, root := range []string{goPath, goCache} {
			_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil || info == nil {
					return nil
				}
				_ = os.Chmod(path, info.Mode()|0o200)
				return nil
			})
		}
	})
	_ = seedHostGoWrapper(t, req.InstallDir, pinGo122)
	req.WithGo = withgo.ResolveOptions{
		InstallDir: req.InstallDir,
		Download:   false,
	}
	ensureDepReplaceHelpersUsed()
	return nil
}

// registerRetryWorkRootCleanup drains WorkRoot before testing.T's TempDir
// RemoveAll (ENOTEMPTY flake under concurrent git/go activity).
func registerRetryWorkRootCleanup(t *testing.T, root string) {
	t.Helper()
	if root == "" {
		return
	}
	t.Cleanup(func() {
		const attempts = 12
		const delay = 15 * time.Millisecond
		var err error
		for i := 0; i < attempts; i++ {
			err = os.RemoveAll(root)
			if err == nil {
				return
			}
			msg := err.Error()
			if !strings.Contains(msg, "directory not empty") && !strings.Contains(msg, "not empty") {
				t.Logf("workroot cleanup: %v", err)
				return
			}
			time.Sleep(delay)
		}
		if err != nil {
			t.Logf("workroot cleanup after %d attempts: %v", attempts, err)
		}
	})
}

// seedHostGoWrapper writes $installDir/<pin>/bin/go that records its arguments,
// GOROOT, and PATH0, then execs the host go so real tidy still works.
func seedHostGoWrapper(t *testing.T, installDir, pin string) string {
	t.Helper()
	hostGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("look up host go: %v", err)
	}
	dest := filepath.Join(installDir, pin)
	bin := filepath.Join(dest, "bin", "go")
	record := filepath.Join(dest, "last-run")
	script := fmt.Sprintf(`#!/bin/sh
printf 'ARGS=%%s\n' "$*" > %q
printf 'GOROOT=%%s\n' "$GOROOT" >> %q
printf 'PATH0=%%s\n' "${PATH%%:*}" >> %q
exec %q "$@"
`, record, record, record, hostGo)
	writeFile(t, bin, script)
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("chmod go wrapper %s: %v", bin, err)
	}
	return record
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

// setupVendorSkip: nearest consumer with require + empty vendor/ (skip tidy).
func setupVendorSkip(t *testing.T, req *Request) {
	t.Helper()
	setupConsumerWithDep(t, req, true)
	vendor := filepath.Join(req.ConsumerModDir, "vendor")
	mkdirAll(t, vendor)
	req.VendorDir = vendor
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
}

// setupConsumerWithDep builds:
//
//	WorkRoot/consumer  (example.com/consumer, optional require)
//	WorkRoot/dep       (example.com/dep)
//
// RepoDir = consumer. requireDep controls D7 fixture.
func setupConsumerWithDep(t *testing.T, req *Request, requireDep bool) {
	t.Helper()
	consumer := filepath.Join(req.WorkRoot, "consumer")
	dep := filepath.Join(req.WorkRoot, "dep")

	body := ""
	if requireDep {
		body = "require " + modDep + " v0.0.1\n"
	}
	writeGoMod(t, consumer, modConsumer, body)
	writeLibPkg(t, consumer, "consumer", "Hello")

	writeGoMod(t, dep, modDep, "")
	writeLibPkg(t, dep, "dep", "Version")

	consumer = resolvePath(t, consumer)
	dep = resolvePath(t, dep)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.DepDir = dep
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modConsumer
	req.WantUpdated = 1
	req.WantCheckouts = 1
	req.WantCheckout = "."
}

// setupConsumerTwoDeps: consumer + dep + dep2 modules.
func setupConsumerTwoDeps(t *testing.T, req *Request) {
	t.Helper()
	setupConsumerWithDep(t, req, true)
	dep2 := filepath.Join(req.WorkRoot, "dep2")
	writeGoMod(t, dep2, modDep2, "")
	writeLibPkg(t, dep2, "dep2", "V2")
	// extend consumer require
	gm := strings.TrimRight(readFile(t, req.ConsumerGoMod), "\n") +
		"\nrequire " + modDep2 + " v0.0.1\n"
	writeFile(t, req.ConsumerGoMod, gm)
	req.Dep2Dir = resolvePath(t, dep2)
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantUpdated = 1
	req.WantCheckouts = 1
	req.WantCheckout = "."
}

// setupNestedConsumerWorkDir: consumer at root; RepoDir = consumer/sub (no go.mod).
// Nearest go.mod walks up to consumer.
func setupNestedConsumerWorkDir(t *testing.T, req *Request) {
	t.Helper()
	setupConsumerWithDep(t, req, true)
	sub := filepath.Join(req.ConsumerModDir, "sub")
	mkdirAll(t, sub)
	req.RepoDir = resolvePath(t, sub)
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantCheckout = ".."
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

// setupDepOnly: dep module exists; no consumer go.mod under WorkRoot (error path).
func setupDepOnly(t *testing.T, req *Request) {
	t.Helper()
	dep := filepath.Join(req.WorkRoot, "dep")
	writeGoMod(t, dep, modDep, "")
	writeLibPkg(t, dep, "dep", "Version")
	req.DepDir = resolvePath(t, dep)
	// RepoDir stays WorkRoot (empty of go.mod)
	req.RepoDir = req.WorkRoot
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

func initStackPrimary(t *testing.T, req *Request) string {
	t.Helper()
	primary := filepath.Join(req.WorkRoot, "primary")
	initGitRepoOnMain(t, primary)
	writeFile(t, filepath.Join(primary, ".gitignore"), "/external\n")
	return primary
}

func seedExternalGitModule(t *testing.T, primary, name, modulePath, goModBody string) string {
	t.Helper()
	dir := filepath.Join(primary, "external", name)
	initGitRepoOnMain(t, dir)
	writeGoMod(t, dir, modulePath, goModBody)
	return resolvePath(t, dir)
}

func seedPlainDep(t *testing.T, req *Request) string {
	t.Helper()
	dep := filepath.Join(req.WorkRoot, "dep")
	writeGoMod(t, dep, modDep, "")
	writeLibPkg(t, dep, "dep", "Version")
	dep = resolvePath(t, dep)
	req.DepDir = dep
	return dep
}

func seedPlainDep2(t *testing.T, req *Request) string {
	t.Helper()
	dep2 := filepath.Join(req.WorkRoot, "dep2")
	writeGoMod(t, dep2, modDep2, "")
	writeLibPkg(t, dep2, "dep2", "V2")
	dep2 = resolvePath(t, dep2)
	req.Dep2Dir = dep2
	return dep2
}

func writeConsumerMainWithImports(t *testing.T, dir string, importPaths ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("package main\n\n")
	if len(importPaths) > 0 {
		b.WriteString("import (\n")
		for _, p := range importPaths {
			fmt.Fprintf(&b, "\t_ %q\n", p)
		}
		b.WriteString(")\n\n")
	}
	b.WriteString("func main() {}\n")
	writeFile(t, filepath.Join(dir, "main.go"), b.String())
}

// setupStackOtherCheckout: primary + external/kool both require dep.
func setupStackOtherCheckout(t *testing.T, req *Request) {
	t.Helper()
	dep := seedPlainDep(t, req)
	_ = dep

	primary := initStackPrimary(t, req)
	kool := seedExternalGitModule(t, primary, "kool", modKool, "require "+modDep+" v0.0.1\n")
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool requirer")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.0.0\n)\n\nreplace %s => ./external/kool\n",
		modDep, modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modKool)
	gitCommitAll(t, primary, "primary + replace kool")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = kool
	req.Consumer2GoMod = filepath.Join(kool, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modKool
	req.WantCheckout = "."
	req.WantCheckout2 = checkoutKool
	req.WantUpdated = 2
	req.WantCheckouts = 2
}

// setupStackSkipNonConsumer: kool has neither require nor replace for dep.
func setupStackSkipNonConsumer(t *testing.T, req *Request) {
	t.Helper()
	_ = seedPlainDep(t, req)

	primary := initStackPrimary(t, req)
	kool := seedExternalGitModule(t, primary, "kool", modKool, "")
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool no require")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.0.0\n)\n\nreplace %s => ./external/kool\n",
		modDep, modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modKool)
	gitCommitAll(t, primary, "primary + replace kool")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = kool
	req.Consumer2GoMod = filepath.Join(kool, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modKool
	req.WantCheckout = "."
	req.WantCheckout2 = checkoutKool
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

// setupStackExistingReplace: kool has replace for dep but no require (still gated).
func setupStackExistingReplace(t *testing.T, req *Request) {
	t.Helper()
	dep := seedPlainDep(t, req)
	oldAbs := filepath.Join(req.WorkRoot, "old-dep-placeholder")
	mkdirAll(t, oldAbs)
	oldAbs = resolvePath(t, oldAbs)

	primary := initStackPrimary(t, req)
	koolBody := fmt.Sprintf("replace %s => %s\n", modDep, oldAbs)
	kool := seedExternalGitModule(t, primary, "kool", modKool, koolBody)
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool existing replace")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.0.0\n)\n\nreplace %s => ./external/kool\n",
		modDep, modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modKool)
	gitCommitAll(t, primary, "primary + replace kool")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = kool
	req.Consumer2GoMod = filepath.Join(kool, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modKool
	req.WantCheckout = "."
	req.WantCheckout2 = checkoutKool
	req.WantUpdated = 2
	req.WantCheckouts = 2
	_ = dep
}

// setupStackSkipSelf: dep checkout lives on the stack via replace; self not rewritten.
func setupStackSkipSelf(t *testing.T, req *Request) {
	t.Helper()
	primary := initStackPrimary(t, req)
	dep := seedExternalGitModule(t, primary, "dep", modDep, "")
	writeLibPkg(t, dep, "dep", "Version")
	gitCommitAll(t, dep, "dep init")
	dep = resolvePath(t, dep)
	req.DepDir = dep

	body := fmt.Sprintf("require %s v0.0.1\n\nreplace %s => ./external/dep\n", modDep, modDep)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep)
	gitCommitAll(t, primary, "primary replace dep")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = dep
	req.Consumer2GoMod = filepath.Join(dep, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modDep
	req.WantCheckout = "."
	req.WantCheckout2 = checkoutDep
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

// setupMultiDirStack: primary requires dep+dep2; kool requires only dep.
func setupMultiDirStack(t *testing.T, req *Request) {
	t.Helper()
	_ = seedPlainDep(t, req)
	_ = seedPlainDep2(t, req)

	primary := initStackPrimary(t, req)
	kool := seedExternalGitModule(t, primary, "kool", modKool, "require "+modDep+" v0.0.1\n")
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool requires dep only")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.0.1\n\t%s v0.0.0\n)\n\nreplace %s => ./external/kool\n",
		modDep, modDep2, modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modDep2, modKool)
	gitCommitAll(t, primary, "primary requires both")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = kool
	req.Consumer2GoMod = filepath.Join(kool, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modKool
	req.WantCheckout = "."
	req.WantCheckout2 = checkoutKool
	req.WantUpdated = 2
	req.WantCheckouts = 2
}

// setupStackZeroGated: git stack whose modules neither require nor replace dep.
func setupStackZeroGated(t *testing.T, req *Request) {
	t.Helper()
	_ = seedPlainDep(t, req)

	primary := initStackPrimary(t, req)
	kool := seedExternalGitModule(t, primary, "kool", modKool, "")
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool empty")

	body := fmt.Sprintf(
		"require %s v0.0.0\n\nreplace %s => ./external/kool\n",
		modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeLibPkg(t, primary, "app", "Hi")
	gitCommitAll(t, primary, "primary no dep gate")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = kool
	req.Consumer2GoMod = filepath.Join(kool, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modKool
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

// assertAbsoluteReplace checks go.mod has replace Old => absolute New.
func assertAbsoluteReplace(t *testing.T, goModPath, oldPath, wantAbs string) {
	t.Helper()
	body := readFile(t, goModPath)
	wantAbs = resolvePath(t, wantAbs)
	foundAbs := false
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "replace ") {
			continue
		}
		if !strings.Contains(trim, oldPath) {
			continue
		}
		parts := strings.Split(trim, "=>")
		if len(parts) != 2 {
			continue
		}
		newPath := strings.TrimSpace(parts[1])
		fields := strings.Fields(newPath)
		if len(fields) > 0 {
			newPath = fields[0]
		}
		// Must be absolute (not ./ or ../ relative).
		if strings.HasPrefix(newPath, "./") || strings.HasPrefix(newPath, "../") {
			t.Fatalf("replace for %s is relative %q; want absolute:\n%s", oldPath, newPath, body)
		}
		resolvedNew := newPath
		if filepath.IsAbs(newPath) {
			if r, err := filepath.EvalSymlinks(newPath); err == nil {
				resolvedNew = r
			}
		} else if !filepath.IsAbs(newPath) && !strings.HasPrefix(newPath, "/") {
			t.Fatalf("replace for %s is not absolute: %q\n%s", oldPath, newPath, body)
		}
		if resolvedNew == wantAbs || newPath == wantAbs || filepath.Clean(newPath) == wantAbs {
			foundAbs = true
			break
		}
		// go may write the path as given; compare cleaned abs forms.
		if filepath.Clean(resolvedNew) == filepath.Clean(wantAbs) {
			foundAbs = true
			break
		}
	}
	if !foundAbs {
		t.Fatalf("expected absolute replace %s => %s in:\n%s", oldPath, wantAbs, body)
	}
}

func assertNoReplaceFor(t *testing.T, goModPath, oldPath string) {
	t.Helper()
	body := readFile(t, goModPath)
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "replace ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trim, "replace "))
		parts := strings.SplitN(rest, "=>", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(strings.TrimSpace(parts[0]))
		if len(left) > 0 && left[0] == oldPath {
			t.Fatalf("did not expect replace for %s:\n%s", oldPath, body)
		}
	}
}

func assertDepReplaceLine(t *testing.T, stdout, modulePath, absDir string) {
	t.Helper()
	// dep-replace <module> => <abs>
	needle := "dep-replace " + modulePath + " => "
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing dep-replace line for %s; got:\n%s", modulePath, stdout)
	}
	// Prefer absolute path token present on the line.
	absDir = resolvePath(t, absDir)
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, needle) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trim, needle))
		if strings.Contains(rest, absDir) || rest == absDir {
			found = true
			break
		}
		// tolerate path without symlink resolve match via Clean
		if filepath.Clean(rest) == filepath.Clean(absDir) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("dep-replace line for %s must include abs %s; stdout:\n%s", modulePath, absDir, stdout)
	}
	// No bare relative-looking NewPath on apply lines
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "dep-replace ") {
			continue
		}
		if strings.Contains(trim, " => ./") || strings.Contains(trim, " => ../") {
			t.Fatalf("apply line must use absolute NewPath, got %q", trim)
		}
	}
}

func assertWouldDepReplaceLine(t *testing.T, stdout, modulePath, absDir string) {
	t.Helper()
	needle := "would: dep-replace " + modulePath + " => "
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing would: dep-replace line for %s; got:\n%s", modulePath, stdout)
	}
	absDir = resolvePath(t, absDir)
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, needle) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trim, needle))
		if strings.Contains(rest, absDir) || filepath.Clean(rest) == filepath.Clean(absDir) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("would: dep-replace line for %s must include abs %s; stdout:\n%s", modulePath, absDir, stdout)
	}
	// Dry-run must not emit bare apply lines.
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "dep-replace ") {
			t.Fatalf("dry-run must not emit bare dep-replace lines: %q", trim)
		}
	}
}

func assertNoTidyArtifacts(t *testing.T, req *Request) {
	t.Helper()
	// Dry-run / vendor-skip: go.sum should not appear if it did not exist.
	if req.ConsumerModDir == "" {
		return
	}
	sum := filepath.Join(req.ConsumerModDir, "go.sum")
	if _, err := os.Stat(sum); err == nil {
		t.Fatalf("go.sum created (tidy must not write): %s", sum)
	}
	if req.Consumer2ModDir != "" {
		sum2 := filepath.Join(req.Consumer2ModDir, "go.sum")
		if _, err := os.Stat(sum2); err == nil {
			t.Fatalf("go.sum created on other checkout (tidy must not write): %s", sum2)
		}
	}
}

func wantCheckoutsOf(req *Request) int {
	if req.WantCheckouts > 0 {
		return req.WantCheckouts
	}
	return 1
}

func checkoutLabelOf(req *Request) string {
	if req.WantCheckout != "" {
		return req.WantCheckout
	}
	return "."
}

func assertApplyBanner(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "==== dep-replace ====") {
		t.Fatalf("stdout missing apply banner ==== dep-replace ====; got:\n%s", stdout)
	}
	if strings.Contains(stdout, "==== dep-replace (dry-run) ====") {
		t.Fatalf("apply must not use dry-run banner; got:\n%s", stdout)
	}
}

func assertDryRunBanner(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "==== dep-replace (dry-run) ====") {
		t.Fatalf("stdout missing dry-run banner ==== dep-replace (dry-run) ====; got:\n%s", stdout)
	}
}

func assertNoBanner(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "====") {
		t.Fatalf("must not print banner; got:\n%s", stdout)
	}
}

func assertDepHeader(t *testing.T, stdout, modulePath string) {
	t.Helper()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "dep  ") && strings.Contains(trim, modulePath) && strings.Contains(trim, "=>") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing dep header %s =>; got:\n%s", modulePath, stdout)
	}
}

func assertReplaceLine(t *testing.T, stdout, modulePath string) {
	t.Helper()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.Contains(trim, "replace  "+modulePath) && strings.Contains(trim, "=>") &&
			!strings.HasPrefix(trim, "would:") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing replace  %s =>; got:\n%s", modulePath, stdout)
	}
}

func assertWouldReplaceLine(t *testing.T, stdout, modulePath string) {
	t.Helper()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "would: replace  ") && strings.Contains(trim, modulePath) && strings.Contains(trim, "=>") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing would: replace  %s =>; got:\n%s", modulePath, stdout)
	}
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace  ") {
			t.Fatalf("dry-run must not emit bare replace lines: %q", trim)
		}
	}
}

func assertDirSummary(t *testing.T, stdout string, modules, checkouts int, dryRun bool) {
	t.Helper()
	var needle string
	if dryRun {
		needle = fmt.Sprintf("dep-replace: would replace in %d modules in %d checkouts", modules, checkouts)
	} else {
		needle = fmt.Sprintf("dep-replace: replaced in %d modules in %d checkouts", modules, checkouts)
	}
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing summary %q; got:\n%s", needle, stdout)
	}
}

func assertGoModUnchangedAt(t *testing.T, path, baseline string) {
	t.Helper()
	if path == "" || baseline == "" {
		t.Fatal("path and baseline required")
	}
	got := readFile(t, path)
	if got != baseline {
		t.Fatalf("go.mod mutated unexpectedly\n got:\n%s\nwant:\n%s", got, baseline)
	}
}

// setupUndoIntroduced: git primary requires dep; HEAD has no replace;
// working tree has absolute replace for dep (simulates --dep-replace / --bring).
// Empty vendor/ skips tidy so offline fixtures need not resolve example.com/*.
func setupUndoIntroduced(t *testing.T, req *Request) {
	t.Helper()
	dep := seedPlainDep(t, req)
	primary := initStackPrimary(t, req)
	writeGoMod(t, primary, modApp, "require "+modDep+" v0.0.1\n")
	writeConsumerMainWithImports(t, primary, modDep)
	mkdirAll(t, filepath.Join(primary, "vendor"))
	gitCommitAll(t, primary, "require dep")

	primary = resolvePath(t, primary)
	gm := strings.TrimRight(readFile(t, filepath.Join(primary, "go.mod")), "\n") +
		"\n\nreplace " + modDep + " => " + dep + "\n"
	writeFile(t, filepath.Join(primary, "go.mod"), gm)

	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.VendorDir = filepath.Join(primary, "vendor")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.WantCheckout = "."
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

// setupUndoKeepsHead: HEAD has replace for kool; WT also introduces dep replace.
// Undo must drop only dep and keep kool. vendor/ skips tidy (offline fixtures).
func setupUndoKeepsHead(t *testing.T, req *Request) {
	t.Helper()
	dep := seedPlainDep(t, req)
	primary := initStackPrimary(t, req)
	kool := seedExternalGitModule(t, primary, "kool", modKool, "")
	writeLibPkg(t, kool, "kool", "Hi")
	gitCommitAll(t, kool, "kool module")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.0.0\n)\n\nreplace %s => ./external/kool\n",
		modDep, modKool, modKool,
	)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modKool)
	mkdirAll(t, filepath.Join(primary, "vendor"))
	gitCommitAll(t, primary, "head has kool replace")

	primary = resolvePath(t, primary)
	gm := strings.TrimRight(readFile(t, filepath.Join(primary, "go.mod")), "\n") +
		"\nreplace " + modDep + " => " + dep + "\n"
	writeFile(t, filepath.Join(primary, "go.mod"), gm)

	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.VendorDir = filepath.Join(primary, "vendor")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.WantCheckout = "."
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

// setupUndoNothing: git primary; WT go.mod matches HEAD (no introduced replace).
func setupUndoNothing(t *testing.T, req *Request) {
	t.Helper()
	_ = seedPlainDep(t, req)
	primary := initStackPrimary(t, req)
	writeGoMod(t, primary, modApp, "require "+modDep+" v0.0.1\n")
	writeConsumerMainWithImports(t, primary, modDep)
	gitCommitAll(t, primary, "clean require")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
}

// setupUndoTwoIntroduced: HEAD clean; WT has replaces for dep and dep2.
// vendor/ skips tidy (offline fixtures).
func setupUndoTwoIntroduced(t *testing.T, req *Request) {
	t.Helper()
	dep := seedPlainDep(t, req)
	dep2 := seedPlainDep2(t, req)
	primary := initStackPrimary(t, req)
	body := fmt.Sprintf("require (\n\t%s v0.0.1\n\t%s v0.0.1\n)\n", modDep, modDep2)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep, modDep2)
	mkdirAll(t, filepath.Join(primary, "vendor"))
	gitCommitAll(t, primary, "require both")

	primary = resolvePath(t, primary)
	gm := strings.TrimRight(readFile(t, filepath.Join(primary, "go.mod")), "\n") +
		"\n\nreplace " + modDep + " => " + dep + "\n" +
		"replace " + modDep2 + " => " + dep2 + "\n"
	writeFile(t, filepath.Join(primary, "go.mod"), gm)

	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.VendorDir = filepath.Join(primary, "vendor")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.WantCheckout = "."
	req.WantUpdated = 1
	req.WantCheckouts = 1
}

func appendAbsoluteReplace(t *testing.T, goModPath, oldPath, absDir string) {
	t.Helper()
	gm := strings.TrimRight(readFile(t, goModPath), "\n") +
		"\nreplace " + oldPath + " => " + absDir + "\n"
	writeFile(t, goModPath, gm)
}

func ensureVendorDir(t *testing.T, modDir string) {
	t.Helper()
	mkdirAll(t, filepath.Join(modDir, "vendor"))
}

func assertHasReplaceFor(t *testing.T, goModPath, oldPath string) {
	t.Helper()
	body := readFile(t, goModPath)
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "replace ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trim, "replace "))
		parts := strings.SplitN(rest, "=>", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(strings.TrimSpace(parts[0]))
		if len(left) > 0 && left[0] == oldPath {
			return
		}
	}
	t.Fatalf("expected replace for %s in:\n%s", oldPath, body)
}

func assertRequireVersion(t *testing.T, goModPath, modulePath, version string) {
	t.Helper()
	body := readFile(t, goModPath)
	needle := modulePath + " " + version
	if !strings.Contains(body, needle) {
		t.Fatalf("expected require %s in:\n%s", needle, body)
	}
}

func assertUndoBanner(t *testing.T, stdout string, dryRun bool) {
	t.Helper()
	want := "==== dep-replace --undo ===="
	if dryRun {
		want = "==== dep-replace --undo (dry-run) ===="
	}
	if !strings.Contains(stdout, want) {
		t.Fatalf("stdout missing banner %q; got:\n%s", want, stdout)
	}
}

func assertDropLine(t *testing.T, stdout, modulePath string, dryRun bool) {
	t.Helper()
	prefix := "drop  " + modulePath + " => "
	if dryRun {
		prefix = "would: drop  " + modulePath + " => "
	}
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(strings.TrimSpace(line), prefix) || strings.HasPrefix(strings.TrimSpace(line), prefix) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing drop line for %s; got:\n%s", modulePath, stdout)
	}
}

func assertUndoSummary(t *testing.T, stdout string, drops, modules, checkouts int, dryRun bool) {
	t.Helper()
	var needle string
	if dryRun {
		needle = fmt.Sprintf("dep-replace: would undo %d replaces in %d modules in %d checkouts", drops, modules, checkouts)
	} else {
		needle = fmt.Sprintf("dep-replace: undid %d replaces in %d modules in %d checkouts", drops, modules, checkouts)
	}
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing summary %q; got:\n%s", needle, stdout)
	}
}

func ensureDepReplaceHelpersUsed() {
	_ = setupConsumerWithDep
	_ = setupConsumerTwoDeps
	_ = setupNestedConsumerWorkDir
	_ = setupDepOnly
	_ = setupVendorSkip
	_ = setupStackOtherCheckout
	_ = setupStackSkipNonConsumer
	_ = setupStackExistingReplace
	_ = setupStackSkipSelf
	_ = setupMultiDirStack
	_ = setupStackZeroGated
	_ = setupUndoIntroduced
	_ = setupUndoKeepsHead
	_ = setupUndoNothing
	_ = setupUndoTwoIntroduced
	_ = appendAbsoluteReplace
	_ = ensureVendorDir
	_ = seedPlainDep
	_ = seedPlainDep2
	_ = initStackPrimary
	_ = seedExternalGitModule
	_ = writeConsumerMainWithImports
	_ = seedHostGoWrapper
	_ = assertAbsoluteReplace
	_ = assertNoReplaceFor
	_ = assertHasReplaceFor
	_ = assertRequireVersion
	_ = assertDepReplaceLine
	_ = assertWouldDepReplaceLine
	_ = assertGoModUnchanged
	_ = assertMutualExclusion
	_ = assertNoTidyArtifacts
	_ = assertApplyBanner
	_ = assertDryRunBanner
	_ = assertUndoBanner
	_ = assertDropLine
	_ = assertUndoSummary
	_ = assertNoBanner
	_ = assertDepHeader
	_ = assertReplaceLine
	_ = assertWouldReplaceLine
	_ = assertDirSummary
	_ = assertGoModUnchangedAt
	_ = wantCheckoutsOf
	_ = checkoutLabelOf
	_ = writeGoMod
	_ = writeLibPkg
}
```
