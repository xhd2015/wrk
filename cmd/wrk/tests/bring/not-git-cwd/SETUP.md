# Scenario

**Feature**: wrk --bring succeeds when consumer cwd is not a git repository

```
# plain non-git cwd + dep git repo -> wrk --bring
#   -> exit 0; external under {abs(cwd)}/external/{basename}-{token}-{date}
#   -> soft-skip replace; do not write .gitignore
#   -> SKIP local dep replacement: <abs-cwd> is not a git repository
# --no-dep: still worktree under external/; no SKIP replace line
# --exec: child runs in external worktree after soft skip
plain dir (no .git) + mydep (git) -> wrk --bring <dep|basename> [--no-dep] [--exec …]
  -> stdout external abs path; worktree owned by dep main
```

## Preconditions

- Git and Go available (parent `bring/`).
- `req.RepoDir` is a plain directory under `WorkRoot` with **no** `.git`.
- Dep path resolves to a real git repo (absolute path or registered basename).
- Parent for external naming is abs(cwd) (`req.ConsumerTop`), not a git toplevel.

## Steps

- Leaves create a plain non-git cwd via `initBringPlainCwd` and a dep via `initBringDepRepo`.
- `req.RepoDir` / `req.ConsumerTop` = plain dir; `req.Args` start with `--bring`.

## Context

- Soft SKIP stderr family matches other bring soft skips (`SKIP local dep replacement: …`).
- No `/external` gitignore write when parent is non-git.
- Basename leaf seeds `WRK_HOME/projects.json` via `wrk --add` (same as `dep/basename-fallback/`).
- Hard-error on non-git for `--dep` / `--all-deps` is **not** retested here.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	ensureBringNotGitCwdHelpersUsed()
	return nil
}

// initBringPlainCwd creates a non-git directory under workRoot and returns its
// symlink-canonical absolute path (matches macOS /var -> /private/var).
func initBringPlainCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	dir := filepath.Join(workRoot, name)
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "README.md"), "# plain non-git cwd\n")
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", dir, err)
	}
	return dir
}

// recordBringSavedProject registers a main repo path in WRK_HOME/projects.json.
func recordBringSavedProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", repoPath)
}

func ensureBringNotGitCwdHelpersUsed() {
	_ = initBringPlainCwd
	_ = recordBringSavedProject
}
```
