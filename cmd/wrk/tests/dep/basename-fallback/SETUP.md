# Scenario

**Feature**: wrk --dep basename fallback to saved projects.json lookup

```
# --dep <basename> missing from cwd -> lookup projects.json by filepath.Base(path)
consumer (git + go.mod) -> wrk --dep <basename> -> stat(cwd/<basename>) fails -> match saved projects

# match count drives outcome (same core as create-mode basename)
0 matches -> wrk: <candidate> does not exist
1 match   -> resolve saved dep path -> external worktree + replace + tidy + gitignore
2+ matches -> TTY numbered prompt OR non-TTY error listing candidates

# fallback skipped
./<basename> exists in cwd (even non-git) -> use cwd path, no lookup
<dir> contains path separator -> no lookup
```

## Preconditions

- Project persistence (`projects.json`, `wrk --add`) is available.
- Consumer repo requires `example.com/dep` and is the process cwd for `wrk --dep`.
- Saved dep repos are registered via `wrk --add` and include a valid Go module.

## Steps

- Descendants configure consumer repo, saved dep paths, `--dep` basename argument, and optional `WRK_BASENAME_CONFIRM` + stdin for TTY selection.
- `req.Args = []string{"--dep", <basename>}` unless the scenario uses a path-with-separator argument.

## Context

- Basename: no path separator, not absolute (`mydep` yes; `sub/mydep`, `/abs`, `../x` no).
- Ambiguous candidates are sorted lexicographically by absolute path before display.
- `WRK_BASENAME_CONFIRM=1` bypasses TTY detection for tests (same pattern as create-mode basename fallback).
- Helpers mirror `projects/basename-fallback/SETUP.md`; dep fixtures reuse `dep/SETUP.md` consumer/dep init.

```go
import (
	"path/filepath"
	"sort"
)

func Setup(t *testing.T, req *Request) error {
	ensureDepBasenameFallbackHelpersUsed()
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

// initSavedDepRepo creates a git dep repo with go.mod at workRoot/parent/basename.
func initSavedDepRepo(t *testing.T, workRoot, parent, basename string) string {
	t.Helper()
	repoPath := filepath.Join(workRoot, parent, basename)
	initGitRepoOnMain(t, repoPath)
	writeFile(t, filepath.Join(repoPath, "go.mod"), "module "+depModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repoPath, "dep.go"), "package dep\n")
	runGitIsolated(t, repoPath, "add", "go.mod", "dep.go")
	runGitIsolated(t, repoPath, "commit", "-m", "add go module")
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

// initConsumerForDepBasename creates a consumer repo requiring example.com/dep.
func initConsumerForDepBasename(t *testing.T, workRoot string) string {
	t.Helper()
	return initConsumerRepo(t, workRoot, true)
}

// initLocalNonGitBasenameInDir creates a non-git directory named basename under dir.
func initLocalNonGitBasenameInDir(t *testing.T, dir, basename string) string {
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

func ensureDepBasenameFallbackHelpersUsed() {
	_ = resolvePath
	_ = initSavedDepRepo
	_ = recordSavedProject
	_ = initConsumerForDepBasename
	_ = initLocalNonGitBasenameInDir
	_ = sortedSavedPaths
}
```