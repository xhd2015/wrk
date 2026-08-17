# Scenario

**Feature**: wrk --dep-update dir-mode fan-out + versioned tidy + `--all` inventory pull

```
# isolated WRK_HOME; consumer go.mod + git-tagged dep / owner modules
# dir mode: pin every existing requirer under git toplevel (else nearest go.mod)
  -> wrk --dep-update <dir>… [--dry-run]
  -> dry-run: would: dep-update; would: go mod tidy | would: skip tidy (vendor/)
  -> apply: drop replace; require@latest; versioned tidy or skip tidy (vendor/)
# --all mode: git toplevel consumers + BuildInventory owners
  -> wrk --dep-update --all [--dry-run]
  -> pin inventory-owned outdated requires; same tidy helper
  -> skip external / same-toplevel filesystem replace; warn no-tag
  -> no dirs / exclusive / bare --all / zero requirers: Error non-zero
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree.
- Go toolchain and **git** on PATH (tag resolution; versioned tidy wrapper execs host `go`).
- Per-leaf `t.TempDir()` WorkRoot; `WRK_HOME={WorkRoot}/.wrk`.
- L2: `req.InProcess = true` (wrkcli.Capture). `Run` passes `CaptureOpts.WithGo`.
- Root Setup seeds `$InstallDir/go1.22.12/bin/go` host-go wrapper (`go 1.22` pin).
  Versioned-tidy leaf also seeds `go1.19.13`. Parallel-safe: no Setenv/Chdir.
- Fixtures seed git tags with `git tag` (isolated git). Existing dir-mode
  consumer dirs stay **non-git** (nearest-only). Fan-out / zero-requirer
  leaves init git at the consumer toplevel.
- Dir-mode apply that tidies and `--all` apply seed `{WorkRoot}/modproxy` +
  `GOPROXY=file://…` so offline tidy resolves synthetic `example.com/*`.
- Parallel-safe: no process env/cwd mutation; inject Env/Dir/`WithGo` via Capture.

## Steps

1. Root `Setup` creates isolated WorkRoot / WrkHome and seeds the `go1.22.12` wrapper.
2. Leaves seed consumer + tagged dep/owner git repos; set Args / paths / extra wrappers.
3. Root `Run` invokes Capture with `WithGo`.

## Context

- **Dir-mode module paths:**
  - `example.com/consumer` — nearest consumer (existing apply leaves; **not** git)
  - `example.com/dep` — primary dep (root-module tags `v0.0.1`, `v0.0.2`)
  - `example.com/dep2` — second dep for multi-dir
  - Fan-out git workspace: `example.com/app` + `example.com/app/pkg` both require dep
  - Skip sibling: `example.com/other` under same toplevel, no require of dep
- **Nested monorepo dep (dir):** git root without go.mod; module at `packages/dep`
  with tags `packages/dep/v0.0.1`, `packages/dep/v0.0.2` → version `v0.0.2`.
- **`--all` module paths:**
  - `example.com/lib` — inventory owner (tags `v1.0.0`, `v1.2.3`)
  - `example.com/app` — consumer under git toplevel (need not be registered)
  - `example.com/external` — non-inventory require (silent skip)
  - `example.com/mono` / `example.com/mono/lib` — same-toplevel local replace skip
  - Nested owner: monorepo owner with `packages/dep` module path
    `example.com/lib/dep`, tags `packages/dep/v0.1.0`, `packages/dep/v0.2.0`
- **Stdout dir apply:** pin line(s) then `go mod tidy ok  module …` or
  `skip tidy  module …  (vendor/)` per consumer. No dir-mode summary.
- **Stdout `--all` apply:** pin lines + tidy/skip + summary
- **Dry-run dir:** `would: dep-update …` + `would: go mod tidy …` or
  `would: skip tidy  module …  (vendor/)`
- **Dry-run `--all`:** `would: dep-update …` + `would: go mod tidy  module …` +
  `dep-update: would update N, already A, skipped S`
- **Tidy pin:** `go 1.22` → wrapper `$InstallDir/go1.22.12/bin/go`; versioned-tidy
  uses `go 1.19` → `go1.19.13`.

```go
import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
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

	modApp      = "example.com/app"
	modAppPkg   = "example.com/app/pkg"
	modOther    = "example.com/other"
	modLib      = "example.com/lib"
	modLibDep   = "example.com/lib/dep"
	modExternal = "example.com/external"
	modMono     = "example.com/mono"
	modMonoLib  = "example.com/mono/lib"

	pinGo122 = "go1.22.12"
	pinGo119 = "go1.19.13"
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
	req.RepoDir = workRoot
	req.InProcess = true
	req.InstallDir = filepath.Join(workRoot, "installed")
	req.ExtraEnv = append(req.ExtraEnv, "HOME="+workRoot)
	_ = seedHostGoWrapper(t, req.InstallDir, pinGo122)
	req.WithGo = withgo.ResolveOptions{
		InstallDir: req.InstallDir,
		Download:   false,
	}
	ensureDepUpdateHelpersUsed()
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

func gitTag(t *testing.T, repo, tag string) {
	t.Helper()
	runGitIsolated(t, repo, "tag", tag)
}

func writeGoMod(t *testing.T, dir, modulePath, body string) {
	t.Helper()
	writeGoModVersion(t, dir, modulePath, "1.22", body)
}

func writeGoModVersion(t *testing.T, dir, modulePath, goVersion, body string) {
	t.Helper()
	mkdirAll(t, dir)
	content := "module " + modulePath + "\n\ngo " + goVersion + "\n"
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

// seedRootTaggedDep creates a git root that is also the module root with tags.
// Returns absolute dep dir. tags like v0.0.1, v0.0.2.
func seedRootTaggedDep(t *testing.T, workRoot, dirName, modulePath string, tags ...string) string {
	t.Helper()
	dep := filepath.Join(workRoot, dirName)
	initGitRepoOnMain(t, dep)
	writeGoMod(t, dep, modulePath, "")
	writeLibPkg(t, dep, "dep", "Version")
	gitCommitAll(t, dep, "dep init")
	for _, tag := range tags {
		gitTag(t, dep, tag)
	}
	return resolvePath(t, dep)
}

// seedNestedPrefixedDep: monorepo git root WITHOUT root go.mod; module at packages/dep.
// Tags packages/dep/v0.0.1, packages/dep/v0.0.2 so CalculateVersionPrefix uses
// filesystem subpath prefix packages/dep/v.
func seedNestedPrefixedDep(t *testing.T, workRoot string) (depDir string, tagHint string) {
	t.Helper()
	repo := filepath.Join(workRoot, "monorepo")
	initGitRepoOnMain(t, repo)
	// Placeholder so root has a commit before submodule files (still no go.mod at root).
	writeFile(t, filepath.Join(repo, "README.md"), "monorepo\n")
	gitCommitAll(t, repo, "root readme")

	dep := filepath.Join(repo, "packages", "dep")
	writeGoMod(t, dep, modDep, "")
	writeLibPkg(t, dep, "dep", "Version")
	gitCommitAll(t, repo, "add packages/dep")
	gitTag(t, repo, "packages/dep/v0.0.1")
	gitTag(t, repo, "packages/dep/v0.0.2")
	return resolvePath(t, dep), "packages/dep/v0.0.2"
}

// writeConsumerWithReplace: consumer go.mod with require@old + absolute replace to dep.
func writeConsumerWithReplace(t *testing.T, req *Request, depDir, modulePath, oldVersion string) {
	t.Helper()
	consumer := filepath.Join(req.WorkRoot, "consumer")
	body := fmt.Sprintf("require %s %s\n\nreplace %s => %s\n",
		modulePath, oldVersion, modulePath, depDir)
	writeGoMod(t, consumer, modConsumer, body)
	writeFile(t, filepath.Join(consumer, "pkg.go"),
		fmt.Sprintf("package consumer\n\nimport _ %q\n\nfunc Hello() string { return \"Hello\" }\n", modulePath))
	consumer = resolvePath(t, consumer)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modConsumer
}

// setupDropReplaceLatest: consumer has replace+require v0.0.1; dep tagged v0.0.1,v0.0.2.
func setupDropReplaceLatest(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	req.DepDir = dep
	req.WantVersion = "v0.0.2"
	writeConsumerWithReplace(t, req, dep, modDep, "v0.0.1")
}

// setupNestedTagPrefix: consumer replace points at packages/dep; tags packages/dep/v*.
func setupNestedTagPrefix(t *testing.T, req *Request) {
	t.Helper()
	dep, tagHint := seedNestedPrefixedDep(t, req.WorkRoot)
	req.DepDir = dep
	req.WantVersion = "v0.0.2"
	req.WantTagHint = tagHint
	writeConsumerWithReplace(t, req, dep, modDep, "v0.0.1")
}

// setupTwoTaggedDeps: consumer replace+require for dep and dep2.
func setupTwoTaggedDeps(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	dep2 := seedRootTaggedDep(t, req.WorkRoot, "dep2", modDep2, "v0.1.0", "v0.1.1")
	req.DepDir = dep
	req.Dep2Dir = dep2
	req.WantVersion = "v0.0.2"
	req.WantVersion2 = "v0.1.1"

	consumer := filepath.Join(req.WorkRoot, "consumer")
	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v0.1.0\n)\n\nreplace %s => %s\nreplace %s => %s\n",
		modDep, modDep2, modDep, dep, modDep2, dep2,
	)
	writeGoMod(t, consumer, modConsumer, body)
	writeFile(t, filepath.Join(consumer, "pkg.go"),
		fmt.Sprintf("package consumer\n\nimport (\n\t_ %q\n\t_ %q\n)\n\nfunc Hello() string { return \"Hello\" }\n",
			modDep, modDep2))
	consumer = resolvePath(t, consumer)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modConsumer
}

// setupNoTagsDep: git dep module with no version tags.
func setupNoTagsDep(t *testing.T, req *Request) {
	t.Helper()
	dep := filepath.Join(req.WorkRoot, "dep")
	initGitRepoOnMain(t, dep)
	writeGoMod(t, dep, modDep, "")
	writeLibPkg(t, dep, "dep", "Version")
	gitCommitAll(t, dep, "dep no tags")
	dep = resolvePath(t, dep)
	req.DepDir = dep
	writeConsumerWithReplace(t, req, dep, modDep, "v0.0.1")
}

// setupDepOnly: valid tagged dep, no consumer go.mod under WorkRoot.
func setupDepOnlyTagged(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.2")
	req.DepDir = dep
	req.RepoDir = req.WorkRoot
	req.WantVersion = "v0.0.2"
}

// --- projects.json + --all fixtures ---

type projectsJSONEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type projectsJSONFile struct {
	Version  int                 `json:"version"`
	Projects []projectsJSONEntry `json:"projects"`
}

// writeProjectsJSON seeds WRK_HOME/projects.json with the given absolute paths.
func writeProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []projectsJSONEntry
	for _, p := range paths {
		projects = append(projects, projectsJSONEntry{
			Path:    p,
			AddedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := projectsJSONFile{Version: 1, Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		t.Fatalf("mkdir WRK_HOME: %v", err)
	}
	path := filepath.Join(wrkHome, "projects.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

// writeConsumerMainWithImports writes main.go that blank-imports each module path
// so go mod tidy keeps those requires.
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

// seedOwnerLibRoot creates owner git repo example.com/lib with tags and a lib package.
func seedOwnerLibRoot(t *testing.T, workRoot string, tags ...string) string {
	t.Helper()
	lib := filepath.Join(workRoot, "repos", "lib")
	initGitRepoOnMain(t, lib)
	writeGoMod(t, lib, modLib, "")
	writeFile(t, filepath.Join(lib, "lib.go"), "package lib\n\nfunc Version() string { return \"ok\" }\n")
	gitCommitAll(t, lib, "init lib")
	for _, tag := range tags {
		gitTag(t, lib, tag)
	}
	return resolvePath(t, lib)
}

// seedAppConsumer creates a git consumer app requiring lib at oldVersion (no replace).
// Consumer is a real git repo so ShowToplevel works for --all.
func seedAppConsumer(t *testing.T, workRoot, oldVersion string, extraRequires ...string) string {
	t.Helper()
	app := filepath.Join(workRoot, "repos", "app")
	initGitRepoOnMain(t, app)
	var b strings.Builder
	b.WriteString("require (\n")
	fmt.Fprintf(&b, "\t%s %s\n", modLib, oldVersion)
	for _, r := range extraRequires {
		parts := strings.SplitN(r, "@", 2)
		if len(parts) != 2 {
			t.Fatalf("extra require %q must be path@version", r)
		}
		fmt.Fprintf(&b, "\t%s %s\n", parts[0], parts[1])
	}
	b.WriteString(")\n")
	writeGoMod(t, app, modApp, b.String())
	writeConsumerMainWithImports(t, app, modLib)
	gitCommitAll(t, app, "init app")
	return resolvePath(t, app)
}

// seedFileModuleProxy publishes srcDir as modulePath@version under a file://
// module proxy root so offline go mod tidy can resolve the version.
func seedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)

	modContent := readFile(t, filepath.Join(srcDir, "go.mod"))
	writeFile(t, filepath.Join(vDir, version+".mod"), modContent)
	writeFile(t, filepath.Join(vDir, version+".info"),
		fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version))
	listPath := filepath.Join(vDir, "list")
	existing := ""
	if data, err := os.ReadFile(listPath); err == nil {
		existing = string(data)
	}
	if !strings.Contains(existing, version) {
		writeFile(t, listPath, existing+version+"\n")
	}

	zipPath := filepath.Join(vDir, version+".zip")
	if err := writeModuleZip(zipPath, modulePath, version, srcDir); err != nil {
		t.Fatalf("write module zip %s: %v", zipPath, err)
	}
}

func writeModuleZip(zipPath, modulePath, version, srcDir string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	prefix := modulePath + "@" + version + "/"
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if path != srcDir && (base == ".git" || base == "sub" || base == "packages") {
				// packages/ skipped only when walking nested multi-module owners
				// that publish a single root module zip; nested-module leaf
				// publishes packages/dep via a dedicated srcDir.
				if base == ".git" {
					return filepath.SkipDir
				}
				if base == "sub" {
					return filepath.SkipDir
				}
			}
			if path != srcDir && base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || rel == ".git" {
			return nil
		}
		rel = filepath.ToSlash(rel)
		w, err := zw.Create(prefix + rel)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, in)
		_ = in.Close()
		return copyErr
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

// enableFileModuleProxy sets ExtraEnv so wrk's child `go` commands resolve
// modules via the local file proxy (offline-friendly tidy).
func enableFileModuleProxy(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	abs, err := filepath.Abs(proxyRoot)
	if err != nil {
		t.Fatalf("abs proxy: %v", err)
	}
	proxyURL := "file://" + abs
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GONOSUMDB=*",
	)
	req.ProxyRoot = abs
}

// setupAllCrossProjectOutdated: owner lib tagged v1.0.0+v1.2.3; app requires v1.0.0.
// Registers owner only (consumer need not be registered). RepoDir = app.
func setupAllCrossProjectOutdated(t *testing.T, req *Request) {
	t.Helper()
	lib := seedOwnerLibRoot(t, req.WorkRoot, "v1.0.0", "v1.2.3")
	app := seedAppConsumer(t, req.WorkRoot, "v1.0.0")
	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = app
	req.WantVersion = "v1.2.3"
	req.WantConsumerModule = modApp
	req.WantUpdated = 1
	req.WantAlready = 0
	req.WantSkipped = 0
	writeProjectsJSON(t, req.WrkHome, lib)
}

// setupAllAlreadyCurrent: same topology; app already at latest tag.
func setupAllAlreadyCurrent(t *testing.T, req *Request) {
	t.Helper()
	lib := seedOwnerLibRoot(t, req.WorkRoot, "v1.0.0", "v1.2.3")
	app := seedAppConsumer(t, req.WorkRoot, "v1.2.3")
	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = app
	req.WantVersion = "v1.2.3"
	req.WantConsumerModule = modApp
	req.WantUpdated = 0
	req.WantAlready = 1
	req.WantSkipped = 0
	writeProjectsJSON(t, req.WrkHome, lib)
}

// setupAllWithProxy seeds cross-project outdated + file proxy for tidy apply leaves.
func setupAllWithProxy(t *testing.T, req *Request) {
	t.Helper()
	setupAllCrossProjectOutdated(t, req)
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, modLib, req.WantVersion, req.OwnerPath)
	// Also publish the old version so any residual resolution succeeds offline.
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.0.0", req.OwnerPath)
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupAllWorktreeOutdated: consumer main + linked worktree; run from worktree.
// Main and worktree start with same outdated require; only worktree must change.
func setupAllWorktreeOutdated(t *testing.T, req *Request) {
	t.Helper()
	lib := seedOwnerLibRoot(t, req.WorkRoot, "v1.0.0", "v1.2.3")
	main := seedAppConsumer(t, req.WorkRoot, "v1.0.0")
	linked := filepath.Join(req.WorkRoot, "repos", "app-wt")
	runGitIsolated(t, main, "worktree", "add", "-b", "feature-dep-update", linked)
	linked = resolvePath(t, linked)
	main = resolvePath(t, main)

	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.MainRepo = main
	req.LinkedWT = linked
	req.ConsumerModDir = linked
	req.ConsumerGoMod = filepath.Join(linked, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	// Snapshot main go.mod separately (must stay unchanged).
	req.RepoDir = linked
	req.WantVersion = "v1.2.3"
	req.WantConsumerModule = modApp
	req.WantUpdated = 1
	req.WantAlready = 0
	req.WantSkipped = 0
	writeProjectsJSON(t, req.WrkHome, lib)

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.2.3", lib)
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.0.0", lib)
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupAllSkipIntraLocalReplace: monorepo with root requiring nested lib via
// filesystem replace (skip) and inventory owner require (bump).
func setupAllSkipIntraLocalReplace(t *testing.T, req *Request) {
	t.Helper()
	lib := seedOwnerLibRoot(t, req.WorkRoot, "v1.0.0", "v1.2.3")

	mono := filepath.Join(req.WorkRoot, "repos", "mono")
	initGitRepoOnMain(t, mono)
	// Nested local module under same toplevel.
	localLib := filepath.Join(mono, "lib")
	writeGoMod(t, localLib, modMonoLib, "")
	writeFile(t, filepath.Join(localLib, "lib.go"), "package lib\n\nfunc Local() string { return \"local\" }\n")

	body := fmt.Sprintf(
		"require (\n\t%s v0.0.1\n\t%s v1.0.0\n)\n\nreplace %s => ./lib\n",
		modMonoLib, modLib, modMonoLib,
	)
	writeGoMod(t, mono, modMono, body)
	writeConsumerMainWithImports(t, mono, modMonoLib, modLib)
	gitCommitAll(t, mono, "init mono")
	mono = resolvePath(t, mono)

	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = mono
	req.ConsumerGoMod = filepath.Join(mono, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = mono
	req.WantVersion = "v1.2.3"
	req.WantConsumerModule = modMono
	// inventory bump for lib; local replace skipped
	req.WantUpdated = 1
	req.WantAlready = 0
	req.WantSkipped = 1
	writeProjectsJSON(t, req.WrkHome, lib)

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.2.3", lib)
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.0.0", lib)
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupAllSkipExternal: app requires external (not inventory) + outdated lib.
func setupAllSkipExternal(t *testing.T, req *Request) {
	t.Helper()
	lib := seedOwnerLibRoot(t, req.WorkRoot, "v1.0.0", "v1.2.3")
	// Use a low major version so go.mod stays path/major valid for tidy.
	app := seedAppConsumer(t, req.WorkRoot, "v1.0.0", modExternal+"@v0.1.0")
	// Blank-import only lib so tidy does not need external published.
	// seedAppConsumer already imports modLib only; external stays as require text.
	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = app
	req.WantVersion = "v1.2.3"
	req.WantConsumerModule = modApp
	req.WantUpdated = 1
	req.WantAlready = 0
	req.WantSkipped = 0 // external is silent, not counted as skipped
	writeProjectsJSON(t, req.WrkHome, lib)

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.2.3", lib)
	seedFileModuleProxy(t, proxyRoot, modLib, "v1.0.0", lib)
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupAllNestedOwnerModuleTag: owner monorepo packages/dep with prefixed tags;
// app requires example.com/lib/dep at older version.
func setupAllNestedOwnerModuleTag(t *testing.T, req *Request) {
	t.Helper()
	owner := filepath.Join(req.WorkRoot, "repos", "lib-mono")
	initGitRepoOnMain(t, owner)
	writeFile(t, filepath.Join(owner, "README.md"), "owner monorepo\n")
	gitCommitAll(t, owner, "root")

	dep := filepath.Join(owner, "packages", "dep")
	writeGoMod(t, dep, modLibDep, "")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n\nfunc V() string { return \"v0.2.0\" }\n")
	gitCommitAll(t, owner, "add packages/dep")
	gitTag(t, owner, "packages/dep/v0.1.0")
	gitTag(t, owner, "packages/dep/v0.2.0")
	owner = resolvePath(t, owner)
	dep = resolvePath(t, dep)

	app := filepath.Join(req.WorkRoot, "repos", "app")
	initGitRepoOnMain(t, app)
	body := fmt.Sprintf("require %s v0.1.0\n", modLibDep)
	writeGoMod(t, app, modApp, body)
	writeConsumerMainWithImports(t, app, modLibDep)
	gitCommitAll(t, app, "init app")
	app = resolvePath(t, app)

	req.OwnerPath = owner
	req.OwnerNestedDir = dep
	req.OwnerGoMod = filepath.Join(dep, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = app
	req.WantVersion = "v0.2.0"
	req.WantTagHint = "packages/dep/v0.2.0"
	req.WantConsumerModule = modApp
	req.WantUpdated = 1
	req.WantAlready = 0
	req.WantSkipped = 0
	writeProjectsJSON(t, req.WrkHome, owner)

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, modLibDep, "v0.2.0", dep)
	seedFileModuleProxy(t, proxyRoot, modLibDep, "v0.1.0", dep)
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupAllNoTagOwner: registered owner with no tags; app requires it.
func setupAllNoTagOwner(t *testing.T, req *Request) {
	t.Helper()
	lib := filepath.Join(req.WorkRoot, "repos", "lib")
	initGitRepoOnMain(t, lib)
	writeGoMod(t, lib, modLib, "")
	writeFile(t, filepath.Join(lib, "lib.go"), "package lib\n\nfunc Version() string { return \"dev\" }\n")
	gitCommitAll(t, lib, "init lib no tags")
	lib = resolvePath(t, lib)

	app := seedAppConsumer(t, req.WorkRoot, "v0.0.1")
	req.OwnerPath = lib
	req.OwnerGoMod = filepath.Join(lib, "go.mod")
	req.BaselineOwnerGoMod = readFile(t, req.OwnerGoMod)
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.RepoDir = app
	req.WantConsumerModule = modApp
	req.WantUpdated = 0
	req.WantAlready = 0
	req.WantSkipped = 1
	writeProjectsJSON(t, req.WrkHome, lib)
}

// seedHostGoWrapper writes $installDir/<pin>/bin/go that records GOROOT/PATH0
// then execs the host go so real `go mod tidy` still works. Returns last-run path.
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
{
  printf 'GOROOT=%%s\n' "$GOROOT"
  IFS=:
  set -- $PATH
  printf 'PATH0=%%s\n' "$1"
} > %q
exec %q "$@"
`, record, hostGo)
	writeFile(t, bin, script)
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("chmod go wrapper %s: %v", bin, err)
	}
	return record
}

// enableDirModeTidyProxy publishes tagged dir-mode deps under file:// GOPROXY.
func enableDirModeTidyProxy(t *testing.T, req *Request) {
	t.Helper()
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	if req.DepDir != "" && req.WantVersion != "" {
		seedFileModuleProxy(t, proxyRoot, modDep, req.WantVersion, req.DepDir)
		seedFileModuleProxy(t, proxyRoot, modDep, "v0.0.1", req.DepDir)
	}
	if req.Dep2Dir != "" && req.WantVersion2 != "" {
		seedFileModuleProxy(t, proxyRoot, modDep2, req.WantVersion2, req.Dep2Dir)
		seedFileModuleProxy(t, proxyRoot, modDep2, "v0.1.0", req.Dep2Dir)
	}
	enableFileModuleProxy(t, req, proxyRoot)
}

func writeRequireReplaceBody(modulePath, oldVersion, depDir string) string {
	return fmt.Sprintf("require %s %s\n\nreplace %s => %s\n",
		modulePath, oldVersion, modulePath, depDir)
}

// setupFanOutRequirers: git workspace; root example.com/app and pkg/ both require dep.
func setupFanOutRequirers(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	req.DepDir = dep
	req.WantVersion = "v0.0.2"

	ws := filepath.Join(req.WorkRoot, "workspace")
	initGitRepoOnMain(t, ws)
	body := writeRequireReplaceBody(modDep, "v0.0.1", dep)
	writeGoMod(t, ws, modApp, body)
	writeConsumerMainWithImports(t, ws, modDep)

	pkg := filepath.Join(ws, "pkg")
	writeGoMod(t, pkg, modAppPkg, body)
	writeConsumerMainWithImports(t, pkg, modDep)
	gitCommitAll(t, ws, "init fan-out consumers")

	ws = resolvePath(t, ws)
	pkg = resolvePath(t, pkg)
	req.RepoDir = ws
	req.ConsumerModDir = ws
	req.ConsumerGoMod = filepath.Join(ws, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = pkg
	req.Consumer2GoMod = filepath.Join(pkg, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modAppPkg
	enableDirModeTidyProxy(t, req)
}

// setupSkipNonRequirer: git workspace with app requiring dep and sibling other that does not.
func setupSkipNonRequirer(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	req.DepDir = dep
	req.WantVersion = "v0.0.2"

	ws := filepath.Join(req.WorkRoot, "workspace")
	initGitRepoOnMain(t, ws)
	app := filepath.Join(ws, "app")
	writeGoMod(t, app, modApp, writeRequireReplaceBody(modDep, "v0.0.1", dep))
	writeConsumerMainWithImports(t, app, modDep)

	other := filepath.Join(ws, "other")
	writeGoMod(t, other, modOther, "")
	writeLibPkg(t, other, "other", "Hi")
	gitCommitAll(t, ws, "init skip-non-requirer")

	ws = resolvePath(t, ws)
	app = resolvePath(t, app)
	other = resolvePath(t, other)
	req.RepoDir = ws
	req.ConsumerModDir = app
	req.ConsumerGoMod = filepath.Join(app, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
	req.Consumer2ModDir = other
	req.Consumer2GoMod = filepath.Join(other, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modOther
	enableDirModeTidyProxy(t, req)
}

// setupVendorSkipDir: nearest consumer (not git) with empty vendor/ beside go.mod.
func setupVendorSkipDir(t *testing.T, req *Request) {
	t.Helper()
	setupDropReplaceLatest(t, req)
	vendor := filepath.Join(req.ConsumerModDir, "vendor")
	mkdirAll(t, vendor)
	req.VendorDir = vendor
}

// setupVersionedTidy: nearest consumer with go 1.19; seed go1.19.13 wrapper.
func setupVersionedTidy(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	req.DepDir = dep
	req.WantVersion = "v0.0.2"

	consumer := filepath.Join(req.WorkRoot, "consumer")
	writeGoModVersion(t, consumer, modConsumer, "1.19", writeRequireReplaceBody(modDep, "v0.0.1", dep))
	writeFile(t, filepath.Join(consumer, "pkg.go"),
		fmt.Sprintf("package consumer\n\nimport _ %q\n\nfunc Hello() string { return \"Hello\" }\n", modDep))
	consumer = resolvePath(t, consumer)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modConsumer
	req.WantGoPin = pinGo119
	req.WrapperRecord = seedHostGoWrapper(t, req.InstallDir, pinGo119)
	enableDirModeTidyProxy(t, req)
}

// setupNoConsumerRequires: git repo whose go.mod does not require the dep.
func setupNoConsumerRequires(t *testing.T, req *Request) {
	t.Helper()
	dep := seedRootTaggedDep(t, req.WorkRoot, "dep", modDep, "v0.0.1", "v0.0.2")
	req.DepDir = dep
	req.WantVersion = "v0.0.2"

	ws := filepath.Join(req.WorkRoot, "workspace")
	initGitRepoOnMain(t, ws)
	writeGoMod(t, ws, modApp, "")
	writeLibPkg(t, ws, "app", "Hi")
	gitCommitAll(t, ws, "init no-requirer")
	ws = resolvePath(t, ws)
	req.RepoDir = ws
	req.ConsumerModDir = ws
	req.ConsumerGoMod = filepath.Join(ws, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp
}

// setupAllVendorSkip: --all cross-project outdated + empty vendor/ (skip tidy).
func setupAllVendorSkip(t *testing.T, req *Request) {
	t.Helper()
	setupAllWithProxy(t, req)
	vendor := filepath.Join(req.ConsumerModDir, "vendor")
	mkdirAll(t, vendor)
	req.VendorDir = vendor
}

// --- asserts ---

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

func assertOwnerGoModUnchanged(t *testing.T, req *Request) {
	t.Helper()
	if req.OwnerGoMod == "" || req.BaselineOwnerGoMod == "" {
		return
	}
	got := readFile(t, req.OwnerGoMod)
	if got != req.BaselineOwnerGoMod {
		t.Fatalf("owner go.mod mutated\n got:\n%s\nwant:\n%s", got, req.BaselineOwnerGoMod)
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

func assertReplacePresentFor(t *testing.T, goModPath, modulePath string) {
	t.Helper()
	body := readFile(t, goModPath)
	found := false
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") && strings.Contains(trim, modulePath) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected replace for %s still present:\n%s", modulePath, body)
	}
}

func assertRequireVersion(t *testing.T, goModPath, modulePath, version string) {
	t.Helper()
	body := readFile(t, goModPath)
	if !strings.Contains(body, modulePath) {
		t.Fatalf("go.mod missing module %s:\n%s", modulePath, body)
	}
	found := false
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "replace ") {
			continue
		}
		if strings.Contains(trim, modulePath) && strings.Contains(trim, version) {
			found = true
			break
		}
	}
	if !found {
		if !strings.Contains(body, modulePath) || !strings.Contains(body, version) {
			t.Fatalf("expected require %s %s in:\n%s", modulePath, version, body)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected require %s %s in:\n%s", modulePath, version, body)
	}
}

func assertDepUpdateLine(t *testing.T, stdout, modulePath, version string) {
	t.Helper()
	needle := "dep-update " + modulePath + " -> " + version
	if !strings.Contains(stdout, needle) {
		alt := "dep-update " + modulePath
		if !strings.Contains(stdout, alt) || !strings.Contains(stdout, version) {
			t.Fatalf("stdout missing dep-update line for %s -> %s; got:\n%s", modulePath, version, stdout)
		}
		found := false
		for _, line := range strings.Split(stdout, "\n") {
			trim := strings.TrimSpace(line)
			if strings.Contains(trim, "dep-update") &&
				strings.Contains(trim, modulePath) &&
				strings.Contains(trim, version) &&
				strings.Contains(trim, "->") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("stdout missing dep-update %s -> %s; got:\n%s", modulePath, version, stdout)
		}
	}
}

func assertWouldDepUpdateLine(t *testing.T, stdout, modulePath, version string) {
	t.Helper()
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "would: dep-update ") {
			continue
		}
		if strings.Contains(trim, modulePath) && strings.Contains(trim, version) && strings.Contains(trim, "->") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing would: dep-update %s -> %s; got:\n%s", modulePath, version, stdout)
	}
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "dep-update ") {
			t.Fatalf("dry-run must not emit bare dep-update lines: %q", trim)
		}
	}
}

func assertNoTidyArtifacts(t *testing.T, req *Request) {
	t.Helper()
	if req.ConsumerModDir == "" {
		return
	}
	sum := filepath.Join(req.ConsumerModDir, "go.sum")
	if _, err := os.Stat(sum); err == nil {
		t.Fatalf("go.sum created (tidy must not run): %s", sum)
	}
}

func assertGoSumExists(t *testing.T, consumerModDir string) {
	t.Helper()
	sum := filepath.Join(consumerModDir, "go.sum")
	if _, err := os.Stat(sum); err != nil {
		t.Fatalf("expected go.sum after tidy at %s: %v", sum, err)
	}
}

func assertTidyOkLine(t *testing.T, stdout, consumerModule string) {
	t.Helper()
	needle := "go mod tidy ok  module " + consumerModule
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing %q; got:\n%s", needle, stdout)
	}
}

func assertWouldTidyLine(t *testing.T, stdout, consumerModule string) {
	t.Helper()
	needle := "would: go mod tidy  module " + consumerModule
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing %q; got:\n%s", needle, stdout)
	}
}

func assertSkipTidyLine(t *testing.T, stdout, consumerModule string) {
	t.Helper()
	needle := "skip tidy  module " + consumerModule + "  (vendor/)"
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing %q; got:\n%s", needle, stdout)
	}
}

func assertWouldSkipTidyLine(t *testing.T, stdout, consumerModule string) {
	t.Helper()
	needle := "would: skip tidy  module " + consumerModule + "  (vendor/)"
	found := false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stdout missing %q; got:\n%s", needle, stdout)
	}
}

func assertVendorUntouched(t *testing.T, vendorDir string) {
	t.Helper()
	if vendorDir == "" {
		t.Fatal("VendorDir required")
	}
	fi, err := os.Stat(vendorDir)
	if err != nil || !fi.IsDir() {
		t.Fatalf("vendor/ dir should remain at %s: %v", vendorDir, err)
	}
	if _, err := os.Stat(filepath.Join(vendorDir, "modules.txt")); err == nil {
		t.Fatalf("vendor/modules.txt created (go mod vendor must not run): %s", vendorDir)
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

func assertVersionedGoUsed(t *testing.T, req *Request) {
	t.Helper()
	if req.WrapperRecord == "" || req.WantGoPin == "" {
		t.Fatal("WrapperRecord and WantGoPin required")
	}
	rec := readFile(t, req.WrapperRecord)
	if !strings.Contains(rec, req.WantGoPin) {
		t.Fatalf("wrapper record missing pin %s:\n%s", req.WantGoPin, rec)
	}
}

func assertAllSummary(t *testing.T, stdout string, updated, already, skipped int, dryRun bool) {
	t.Helper()
	var needle string
	if dryRun {
		needle = fmt.Sprintf("dep-update: would update %d, already %d, skipped %d", updated, already, skipped)
	} else {
		needle = fmt.Sprintf("dep-update: updated %d, already %d, skipped %d", updated, already, skipped)
	}
	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout missing summary %q; got:\n%s", needle, stdout)
	}
}

func assertAlreadyUpToDateBanner(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "dep-update: already up to date") {
		t.Fatalf("stdout missing already-up-to-date banner; got:\n%s", stdout)
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout empty; expected trailing newline content")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing \\n; got %q", stdout)
	}
}

func assertWarningStderr(t *testing.T, stderr string) {
	t.Helper()
	lower := strings.ToLower(stderr)
	if !strings.Contains(lower, "warning:") && !strings.Contains(lower, "warning") {
		t.Fatalf("stderr should contain warning: prefix, got %q", stderr)
	}
}

func ensureDepUpdateHelpersUsed() {
	_ = setupDropReplaceLatest
	_ = setupNestedTagPrefix
	_ = setupTwoTaggedDeps
	_ = setupNoTagsDep
	_ = setupDepOnlyTagged
	_ = seedRootTaggedDep
	_ = seedNestedPrefixedDep
	_ = writeConsumerWithReplace
	_ = writeProjectsJSON
	_ = seedOwnerLibRoot
	_ = seedAppConsumer
	_ = seedFileModuleProxy
	_ = enableFileModuleProxy
	_ = setupAllCrossProjectOutdated
	_ = setupAllAlreadyCurrent
	_ = setupAllWithProxy
	_ = setupAllWorktreeOutdated
	_ = setupAllSkipIntraLocalReplace
	_ = setupAllSkipExternal
	_ = setupAllNestedOwnerModuleTag
	_ = setupAllNoTagOwner
	_ = writeConsumerMainWithImports
	_ = assertNoReplaceFor
	_ = assertReplacePresentFor
	_ = assertRequireVersion
	_ = assertDepUpdateLine
	_ = assertWouldDepUpdateLine
	_ = assertGoModUnchanged
	_ = assertOwnerGoModUnchanged
	_ = assertMutualExclusion
	_ = assertNoTidyArtifacts
	_ = assertGoSumExists
	_ = assertTidyOkLine
	_ = assertWouldTidyLine
	_ = assertAllSummary
	_ = assertAlreadyUpToDateBanner
	_ = assertStdoutTrailingNewline
	_ = assertWarningStderr
	_ = writeGoMod
	_ = writeGoModVersion
	_ = writeLibPkg
	_ = initGitRepoOnMain
	_ = gitCommitAll
	_ = gitTag
	_ = seedHostGoWrapper
	_ = enableDirModeTidyProxy
	_ = writeRequireReplaceBody
	_ = setupFanOutRequirers
	_ = setupSkipNonRequirer
	_ = setupVendorSkipDir
	_ = setupVersionedTidy
	_ = setupNoConsumerRequires
	_ = setupAllVendorSkip
	_ = assertSkipTidyLine
	_ = assertWouldSkipTidyLine
	_ = assertVendorUntouched
	_ = assertGoModUnchangedAt
	_ = assertVersionedGoUsed
}
```
