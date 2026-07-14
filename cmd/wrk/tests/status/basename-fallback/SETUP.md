# Scenario

**Feature**: wrk --status basename fallback to saved projects.json lookup

```
# wrk <basename> --status missing from cwd -> lookup projects.json by filepath.Base(path)
neutral cwd -> wrk myrepo --status -> stat(cwd/myrepo) fails -> match saved projects -> status blocks

# match count drives outcome (same core as create-mode basename)
0 matches -> wrk: <candidate> does not exist
1 match   -> resolve saved path -> scan_repo status blocks for saved repo root
2+ matches -> TTY numbered prompt OR non-TTY error listing candidates

# fallback skipped
./<basename> exists in cwd (even non-git) -> use cwd path, no lookup
<dir> contains path separator -> no lookup
```

## Preconditions

- Project persistence (`projects.json`, `wrk --add`) is available.
- `wrk --status` is active (`req.Args = []string{"--status"}` from parent `status/SETUP.md`).
- Cwd for basename tests is a neutral directory without a matching local entry unless the scenario requires one.

## Steps

- Descendants configure saved project paths, cwd, `<dir>` basename argument, and optional `WRK_BASENAME_CONFIRM` + stdin for TTY selection.
- Command form: `wrk <basename> --status` via `req.TargetDir` + inherited `req.Args`.

## Context

- Basename: no path separator, not absolute (`myrepo` yes; `saved/myrepo`, `/abs`, `../x` no).
- Ambiguous candidates are sorted lexicographically by absolute path before display.
- `WRK_BASENAME_CONFIRM=1` bypasses TTY detection for tests (same pattern as create-mode basename fallback).
- Helpers mirror `projects/basename-fallback/SETUP.md`; status output helpers come from `status/SETUP.md`.

```go
import (
	"path/filepath"
	"sort"
)

func Setup(t *testing.T, req *Request) error {
	ensureStatusBasenameFallbackHelpersUsed()
	return nil
}

// resolvePath returns an absolute, symlink-canonicalized path for assertions.
func resolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

// initSavedGitRepo creates a git repo at workRoot/parent/basename and returns its path.
func initSavedGitRepo(t *testing.T, workRoot, parent, basename string) string {
	t.Helper()
	repoPath := filepath.Join(workRoot, parent, basename)
	initGitRepoOnMain(t, repoPath)
	repoPath, err := filepath.EvalSymlinks(repoPath)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", repoPath, err)
	}
	return repoPath
}

// recordSavedProject registers a main repo in projects.json via wrk --add.
func recordSavedProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, req.WorkRoot, "--add", repoPath)
}

// initNeutralCwd creates an empty non-git directory for wrk invocations.
func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	cwd := filepath.Join(workRoot, name)
	mkdirAll(t, cwd)
	return cwd
}

// initNonGitBasenameDir creates a non-git directory named basename directly under dir.
func initNonGitBasenameInDir(t *testing.T, dir, basename string) string {
	t.Helper()
	path := filepath.Join(dir, basename)
	mkdirAll(t, path)
	return path
}

// sortedSavedPaths returns lexicographically sorted absolute paths.
func sortedSavedPaths(t *testing.T, paths ...string) []string {
	t.Helper()
	var out []string
	for _, p := range paths {
		out = append(out, resolvePath(t, p))
	}
	sort.Strings(out)
	return out
}

func ensureStatusBasenameFallbackHelpersUsed() {
	_ = resolvePath
	_ = initSavedGitRepo
	_ = recordSavedProject
	_ = initNeutralCwd
	_ = initNonGitBasenameInDir
	_ = sortedSavedPaths
}
```
