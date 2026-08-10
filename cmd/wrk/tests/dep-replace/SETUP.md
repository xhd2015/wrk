# Scenario

**Feature**: wrk --dep-replace writes absolute replace into nearest consumer go.mod

```
# isolated WRK_HOME + plain go.mod fixtures under WorkRoot (git not required)
consumer go.mod + dep module dir(s)
  -> wrk --dep-replace <dir>… [--dry-run]
  -> dry-run: would: dep-replace <mod> => <abs>; no go.mod write; no tidy
  -> apply: dep-replace <mod> => <abs>; absolute replace; no tidy
  -> multi-arg fail-fast; exclusive / empty / missing / not-module: Error non-zero
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- Go toolchain on PATH (uses `go mod edit` via product/library).
- Per-leaf `t.TempDir()` WorkRoot; `WRK_HOME={WorkRoot}/.wrk`.
- L2: every leaf sets `req.InProcess = true` (wrkcli.Capture).
- **No git required** for replace success paths (unlike `--dep-update` / pin-locals stack).
- Offline-friendly: no network needed (`go mod edit` only; no tidy).

## Steps

1. Root `Setup` creates isolated `WorkRoot` / `WrkHome`.
2. Leaves seed consumer + dep go.mod fixtures and set `Args` / `RepoDir` / paths.
3. Root `Run` invokes wrk via Capture when `InProcess`.

## Context

- **Module paths** used in fixtures:
  - `example.com/consumer` — nearest consumer
  - `example.com/dep` — primary dep
  - `example.com/dep2` — second dep (multi-dir)
- **Stdout apply line:** `dep-replace <module> => <abs>`
- **Dry-run:** prefix `would: `
- **Consumer resolution:** walk up from `RepoDir` (Capture Dir) to nearest go.mod.
- Absolute NewPath must match resolved dep dir (EvalSymlinks-normalized when needed).

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
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
	ensureDepReplaceHelpersUsed()
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
		if strings.HasPrefix(trim, "replace ") && strings.Contains(trim, oldPath) {
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
	// D2: no tidy — go.sum should not appear if it did not exist.
	if req.ConsumerModDir == "" {
		return
	}
	sum := filepath.Join(req.ConsumerModDir, "go.sum")
	if _, err := os.Stat(sum); err == nil {
		// Allowed only if baseline already had go.sum (we never create one).
		t.Fatalf("go.sum created (tidy must not run): %s", sum)
	}
}

func ensureDepReplaceHelpersUsed() {
	_ = setupConsumerWithDep
	_ = setupConsumerTwoDeps
	_ = setupNestedConsumerWorkDir
	_ = setupDepOnly
	_ = assertAbsoluteReplace
	_ = assertNoReplaceFor
	_ = assertDepReplaceLine
	_ = assertWouldDepReplaceLine
	_ = assertGoModUnchanged
	_ = assertMutualExclusion
	_ = assertNoTidyArtifacts
	_ = writeGoMod
	_ = writeLibPkg
}
```
