# Scenario

**Feature**: repeatable `wrk --bring p1 --bring p2` brings multiple deps in one invocation

```
# Multi-bring (v1): lessflags.StringSlice --bring only (not multi-value sugar)
# Order left → right; stdout one abs path per line (trailing \n)
# Soft SKIP continues; hard errors fail-fast (keep earlier external WTs)
# --exec rejected when len(bringPaths) > 1
# Positionals with --bring rejected (unexpected arguments)
# Exact duplicate bring args → error
# --no-dep applies to all brought deps
# P2 preflight-ambiguous: multi ambiguous basenames Select one-by-one;
#   after all resolves succeed, stderr "will bring:" plan (multi-only)
consumer + dep1 + dep2 -> wrk --bring <dep1> --bring <dep2>
  -> external/{basename}-main-{date} per dep
  -> replace/tidy best-effort per dep (skipped with --no-dep)
```

## Preconditions

- Same as parent `bring/`: Git + Go on PATH.
- L2 only: every leaf sets `req.InProcess = true`.
- Do **not** redefine root `Request` / `Response` / `Run`. Reuse parent helpers
  (`initBringConsumerRepo`, `initBringDepRepo`, `bringExternalWorktreePath`, …).

## Steps

- Leaves build consumer + one or two dep repos under `req.WorkRoot`.
- Field mapping:
  - `req.DepPath` — first dep main path (or sole dep)
  - `req.SecondRepo` — second dep main path when needed
  - `req.ConsumerTop` / `req.RepoDir` — consumer git toplevel / cwd
  - `req.ExternalWtDir` / `req.ExternalWtDir2` — expected external paths (assert may set)
- `req.Args` uses repeated `--bring` flags, e.g. `{"--bring", dep1, "--bring", dep2}`.

## Context

- Locked CLI form: **repeatable** `--bring` only (`wrk --bring p1 --bring p2`).
  Not v1: `wrk --bring p1 p2` multi-value sugar.
- Preferred hard-error prefixes: `wrk:`; soft SKIP uses existing
  `SKIP local dep replacement: …` wording.
- Preferred multi+exec error: `wrk: --exec is only valid with a single --bring path`
  (asserts accept stable equivalents mentioning `--exec` + single/one/multiple).
- Existing single-bring coverage remains under `basic/` etc.; `single-compat/`
  proves one `--bring` still works when the flag becomes a slice.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

const (
	multiBringDep1Module = "example.com/dep1"
	multiBringDep2Module = "example.com/dep2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	// Parent bring Setup already skipIfNoGit + go check.
	ensureMultiBringHelpersUsed()
	return nil
}

// initMultiBringConsumerWithTwoRequires creates consumer requiring dep1+dep2 modules.
// Same require-only pattern as other bring leaves (no source imports; avoid network tidy).
// Implementer should replace all brought deps before a final tidy (or re-pin requires)
// so the second require is not dropped mid-loop.
func initMultiBringConsumerWithTwoRequires(t *testing.T, workRoot string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runBringGo(t, consumer, "mod", "edit", "-require="+multiBringDep1Module+"@v0.0.0")
	runBringGo(t, consumer, "mod", "edit", "-require="+multiBringDep2Module+"@v0.0.0")
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks consumer: %v", err)
	}
	return consumer
}

// initMultiBringDepRepo creates a named dep repo with a distinct module path.
func initMultiBringDepRepo(t *testing.T, workRoot, name, modulePath string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	runGitIsolated(t, dep, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep, "commit", "-m", "add "+name+" module")
	dep, err := filepath.EvalSymlinks(dep)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", dep, err)
	}
	return dep
}

// multiCountExternalDirs counts direct children of consumerTop/external.
func multiCountExternalDirs(t *testing.T, consumerTop string) int {
	t.Helper()
	ext := filepath.Join(consumerTop, "external")
	entries, err := os.ReadDir(ext)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("readdir %s: %v", ext, err)
	}
	return len(entries)
}

// assertStdoutTwoPathsExact asserts stdout is path1\npath2\n (order preserved).
func assertStdoutTwoPathsExact(t *testing.T, stdout, path1, path2 string) {
	t.Helper()
	body := path1 + "\n" + path2 + "\n"
	assert.Output(t, stdout, v2StdoutTemplate(body))
}

// multiSnapshotBringGoMod writes consumer go.mod bytes for later compare (--no-dep).
func multiSnapshotBringGoMod(t *testing.T, req *Request, modDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("snapshot go.mod: %v", err)
	}
	path := filepath.Join(req.WorkRoot, "go.mod.before")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write go.mod.before: %v", err)
	}
}

// multiAssertBringGoModUnchanged fails if go.mod differs from multiSnapshotBringGoMod.
func multiAssertBringGoModUnchanged(t *testing.T, req *Request, modDir string) {
	t.Helper()
	before, err := os.ReadFile(filepath.Join(req.WorkRoot, "go.mod.before"))
	if err != nil {
		t.Fatalf("read go.mod.before: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("go.mod changed under multi --no-dep\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// multiRecordSavedProject registers a main repo via wrk --add (basename multi leaves).
func multiRecordSavedProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", repoPath)
}

func ensureMultiBringHelpersUsed() {
	_ = initMultiBringConsumerWithTwoRequires
	_ = initMultiBringDepRepo
	_ = multiCountExternalDirs
	_ = assertStdoutTwoPathsExact
	_ = multiSnapshotBringGoMod
	_ = multiAssertBringGoModUnchanged
	_ = multiRecordSavedProject
	_ = multiBringDep1Module
	_ = multiBringDep2Module
}
```
