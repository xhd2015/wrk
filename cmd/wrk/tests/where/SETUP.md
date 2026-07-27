# Scenario

**Feature**: wrk --where basename lookup in projects.json

```
# read-only lookup by basename only (no cwd stat, no disk path resolution)
wrk --where <basename> -> FindProjectsByBasename(projects.json)

# match count drives stdout
0 matches -> non-zero exit, stderr no-match message, empty stdout
1 match   -> exit 0, stdout = one absolute path + trailing newline
2+ matches -> exit 0, stdout = all paths sorted lexicographically, one per line

# basename-only input
path separator or absolute path -> non-zero exit, basename-only rejection

# standalone mode
mutually exclusive with other modes; no extra positionals
```

## Preconditions

- Project persistence (`projects.json`, `wrk --add`) is available.
- Tests seed saved projects via `wrk --add`.
- Lookup uses basename `spl` unless a descendant overrides `WhereBasename`.

## Steps

- Descendants configure saved project paths, cwd, and `req.Args = []string{"--where", <basename>}`.
- Cwd is a neutral directory unless the scenario requires a local `./spl` entry.

## Context

- Basename: no path separator, not absolute (`spl` yes; `sub/spl`, `/abs/spl`, `../spl` no).
- Unlike create-mode basename fallback, `--where` never stats cwd or resolves paths on disk.
- Multi-match prints all paths (no TTY prompt); candidates sorted lexicographically.

```go
import (
	"path/filepath"
	"sort"
	"github.com/xhd2015/doctest/session"
)

const whereBasename = "spl"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
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

// initNonGitBasenameDir creates a non-git directory named basename directly under dir.
func initNonGitBasenameInDir(t *testing.T, dir, basename string) string {
	t.Helper()
	path := filepath.Join(dir, basename)
	mkdirAll(t, path)
	return path
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

// whereArgs returns CLI args for wrk --where <basename>.
func whereArgs(basename string) []string {
	return []string{"--where", basename}
}

func ensureWhereHelpersUsed() {
	_ = initSavedGitRepo
	_ = recordSavedProject
	_ = initNeutralCwd
	_ = initNonGitBasenameInDir
	_ = resolvePath
	_ = sortedSavedPaths
	_ = whereArgs
}```
