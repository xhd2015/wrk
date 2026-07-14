# Scenario

**Feature**: wrk basename fallback to saved projects (create mode only)

```
# basename missing from cwd -> lookup projects.json by filepath.Base(path)
wrk <basename> (create mode) -> stat(cwd/<basename>) fails -> match saved projects

# match count drives outcome
0 matches -> wrk: <candidate> does not exist
1 match   -> create worktree from saved project path
2+ matches -> TTY numbered prompt OR non-TTY error listing candidates

# fallback skipped
./<basename> exists in cwd (even non-git) -> use cwd path, no lookup
<dir> contains path separator -> no lookup
--done / other non-create modes (except --list, --status, --repos) -> no lookup
```

## Preconditions

- Project persistence (`projects.json`, `wrk --add`) is available.
- Tests seed saved projects via `wrk --add` or pre-populated `projects.json`.
- Cwd for basename tests is a neutral directory without a matching local entry unless the scenario requires one.

## Steps

- Descendants configure saved project paths, cwd, `<dir>` basename argument, and mode flags.
- TTY selection tests set `WRK_BASENAME_CONFIRM=1` and pipe the selection index on stdin.

## Context

- Basename: no path separator, not absolute (`myrepo` yes; `sub/foo`, `/abs`, `../x` no).
- Ambiguous candidates are sorted lexicographically by absolute path before display.
- `WRK_BASENAME_CONFIRM=1` bypasses TTY detection for tests (same pattern as `WRK_SET_TASK_CONFIRM`).

```go
import (
	"path/filepath"
	"sort"
)

func Setup(t *testing.T, req *Request) error {
	ensureBasenameFallbackHelpersUsed()
	return nil
}

// initSavedGitRepo creates a git repo at workRoot/parent/basename and returns its path.
func initSavedGitRepo(t *testing.T, workRoot, parent, basename string) string {
	t.Helper()
	repoPath := filepath.Join(workRoot, parent, basename)
	initGitRepoOnMain(t, repoPath)
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

// initNonGitBasenameDir creates a non-git directory named basename directly under workRoot.
func initNonGitBasenameDir(t *testing.T, workRoot, basename string) string {
	t.Helper()
	path := filepath.Join(workRoot, basename)
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

func ensureBasenameFallbackHelpersUsed() {
	_ = initSavedGitRepo
	_ = recordSavedProject
	_ = initNeutralCwd
	_ = initNonGitBasenameDir
	_ = sortedSavedPaths
}
```