# Scenario

**Feature**: wrk --all-deps uses registered projects.json entries as the sole dep-discovery source

```
# ListProjects(wrkHome) yields sorted main-repo paths; each is scanned with mod/scan
WRK_HOME/projects.json -> ListProjects -> mod/scan per project -> match required modules -> link external worktrees

# skip rules preserved
self / not-required / already-replaced / seen dedup / non-git / missing path -> skip silently
```

## Preconditions

- Dep repos are registered in `{WRK_HOME}/projects.json` via `wrk --add` or direct seeding.
- No `--scan-root` or `WRK_SCAN_ROOT`; discovery is projects-only.

## Steps

- Descendants seed `projects.json`, create consumer + dep fixtures, and run `wrk --all-deps`.
- Stdout order follows lexicographic registered project path order; within a project, modules follow `mod/scan` Dir order.

## Context

- One external worktree per matched repo (multiple sub-modules → multiple replaces, one worktree).
- Empty or absent `projects.json` → `wrked 0 deps` with no side effects.
- Branch per dep is `{token}-{date}[-N]` without dep basename (P2). Separate dep
  repos both on `main` each get `main-{date}` (collision is per depMain; do **not**
  force artificial `-1` across repos). Preferred-branch pre-exists in one dep →
  `-1` is covered by `dep/branch-collision-suffix/`. See `multi-dep-branch-names/`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	allDepsEnsureHelpersUsed()
	return nil
}

// initNestedSubmoduleRepo creates a git repo whose root module is example.com/myrepo
// and whose nested sub-module at services/dep is example.com/dep.
func initNestedSubmoduleRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "go.mod"), "module example.com/myrepo\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "root.go"), "package myrepo\n")
	depDir := filepath.Join(path, "services", "dep")
	mkdirAll(t, depDir)
	writeFile(t, filepath.Join(depDir, "go.mod"), "module example.com/dep\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depDir, "dep.go"), "package dep\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init myrepo with nested dep")
}

// initMultiModuleRepo creates a git repo with NO root go.mod and two nested
// sub-modules: svc-a (example.com/dep1) and svc-b (example.com/dep2).
func initMultiModuleRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	mkdirAll(t, filepath.Join(path, "svc-a"))
	mkdirAll(t, filepath.Join(path, "svc-b"))
	writeFile(t, filepath.Join(path, "svc-a", "go.mod"), "module example.com/dep1\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "svc-a", "a.go"), "package dep1\n")
	writeFile(t, filepath.Join(path, "svc-b", "go.mod"), "module example.com/dep2\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "svc-b", "b.go"), "package dep2\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init myrepo with two sub-modules")
}

func registeredEnsureHelpersUsed() {
	_ = initNestedSubmoduleRepo
	_ = initMultiModuleRepo
}
```