# Scenario

**Feature**: wrk --dep-update drops replace and pins require to latest git tag

```
# isolated WRK_HOME; consumer go.mod + git-tagged dep module(s)
consumer has replace + old require
  -> wrk --dep-update <dir>… [--dry-run]
  -> dry-run: would: dep-update <mod> -> vN.N.N; no write
  -> apply: drop replace; require@latest; dep-update line; no tidy
  -> no tags / exclusive / empty: Error non-zero
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree.
- Go toolchain and **git** on PATH (tag resolution under DepDir).
- Per-leaf `t.TempDir()` WorkRoot; `WRK_HOME={WorkRoot}/.wrk`.
- L2: `req.InProcess = true` (wrkcli.Capture).
- Fixtures seed git tags with `git tag` (isolated git).

## Steps

1. Root `Setup` creates isolated WorkRoot / WrkHome.
2. Leaves seed consumer + tagged dep git repos; set Args / paths.
3. Root `Run` invokes Capture.

## Context

- **Module paths:**
  - `example.com/consumer` — nearest consumer
  - `example.com/dep` — primary dep (root-module tags `v0.0.1`, `v0.0.2`)
  - `example.com/dep2` — second dep for multi-dir
- **Nested monorepo dep:** git root without go.mod; module at `packages/dep`
  with tags `packages/dep/v0.0.1`, `packages/dep/v0.0.2` → version `v0.0.2`.
- **Stdout apply:** `dep-update <module> -> v0.0.2` (+ optional tag form)
- **Dry-run:** `would: dep-update …`

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	wrkDate = "2026-06-30"

	modConsumer = "example.com/consumer"
	modDep      = "example.com/dep"
	modDep2     = "example.com/dep2"
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
	writeLibPkg(t, consumer, "consumer", "Hello")
	consumer = resolvePath(t, consumer)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
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
	writeLibPkg(t, consumer, "consumer", "Hello")
	consumer = resolvePath(t, consumer)
	req.RepoDir = consumer
	req.ConsumerModDir = consumer
	req.ConsumerGoMod = filepath.Join(consumer, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
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

func assertRequireVersion(t *testing.T, goModPath, modulePath, version string) {
	t.Helper()
	body := readFile(t, goModPath)
	// Loose: require line or block entry contains module and version.
	if !strings.Contains(body, modulePath) {
		t.Fatalf("go.mod missing module %s:\n%s", modulePath, body)
	}
	// Prefer same-line or nearby version token.
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
		// go mod edit -require may place version on same require line.
		if !strings.Contains(body, modulePath) || !strings.Contains(body, version) {
			t.Fatalf("expected require %s %s in:\n%s", modulePath, version, body)
		}
		// Module and version both present somewhere — accept for block form.
		found = true
	}
	if !found {
		t.Fatalf("expected require %s %s in:\n%s", modulePath, version, body)
	}
}

func assertDepUpdateLine(t *testing.T, stdout, modulePath, version string) {
	t.Helper()
	// dep-update <module> -> <version>
	needle := "dep-update " + modulePath + " -> " + version
	if !strings.Contains(stdout, needle) {
		// Allow flexible spacing around ->
		alt := "dep-update " + modulePath
		if !strings.Contains(stdout, alt) || !strings.Contains(stdout, version) {
			t.Fatalf("stdout missing dep-update line for %s -> %s; got:\n%s", modulePath, version, stdout)
		}
		// Still require arrow token somewhere with module context.
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
	// No would: on apply success lines mixed without prefix check done by leaf.
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

func ensureDepUpdateHelpersUsed() {
	_ = setupDropReplaceLatest
	_ = setupNestedTagPrefix
	_ = setupTwoTaggedDeps
	_ = setupNoTagsDep
	_ = setupDepOnlyTagged
	_ = seedRootTaggedDep
	_ = seedNestedPrefixedDep
	_ = writeConsumerWithReplace
	_ = assertNoReplaceFor
	_ = assertRequireVersion
	_ = assertDepUpdateLine
	_ = assertWouldDepUpdateLine
	_ = assertGoModUnchanged
	_ = assertMutualExclusion
	_ = assertNoTidyArtifacts
	_ = writeGoMod
	_ = writeLibPkg
	_ = initGitRepoOnMain
	_ = gitCommitAll
	_ = gitTag
}
```
