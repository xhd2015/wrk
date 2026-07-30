# Scenario

**Feature**: wrk --where basename lookup in projects.json (Bool flag + positional)

```
# Bool("--where") + exactly one basename positional (either order)
wrk --where <basename>  OR  wrk <basename> --where
  -> FindProjectsByBasename(projects.json)

# match count drives stdout
0 matches -> non-zero exit, stderr no-match message, empty stdout
1 match   -> exit 0, stdout = one absolute path + trailing newline
2+ matches -> exit 0, stdout = all paths sorted lexicographically, one per line

# basename-only input
path separator or absolute path -> non-zero exit, basename-only rejection

# binding / arity
wrk --where              -> wrk: --where requires a path argument
wrk --where=spl          -> fail (equals form; no treat-as-basename)
wrk --where --main       -> compose: print main of cwd (see main-mode/compose/where)

# standalone mode
mutually exclusive with other modes; no extra positionals beyond the one basename
```

## Preconditions

- Project persistence (`projects.json`, `wrk --add`) is available.
- Tests seed saved projects via `wrk --add`.
- Lookup uses basename `spl` unless a descendant overrides `WhereBasename`.
- `--where` is a **Bool** flag; operand is a remaining positional (breaking: was String-bound).

## Steps

- Descendants configure saved project paths, cwd, and CLI form
  (`req.Args = []string{"--where", <basename>}` or basename-then-flag via `TargetDir`).
- Cwd is a neutral directory unless the scenario requires a local `./spl` entry.

## Context

- Basename: no path separator, not absolute (`spl` yes; `sub/spl`, `/abs/spl`, `../spl` no).
- Unlike create-mode basename fallback, `--where` never stats cwd or resolves paths on disk.
- Multi-match prints all paths (no TTY prompt); candidates sorted lexicographically.
- Operand-then-flag mirrors `--cd`: `wrk <basename> --where` is valid.

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

// whereArgs returns CLI args for wrk --where <basename> (flag then basename).
func whereArgs(basename string) []string {
	return []string{"--where", basename}
}

// setWhereFlagThenBasename: wrk --where <basename>
func setWhereFlagThenBasename(req *Request, basename string) {
	req.Args = []string{"--where", basename}
	req.TargetDir = ""
}

// setWhereBasenameThenFlag: wrk <basename> --where
func setWhereBasenameThenFlag(req *Request, basename string) {
	req.TargetDir = basename
	req.Args = []string{"--where"}
}

func ensureWhereHelpersUsed() {
	_ = initSavedGitRepo
	_ = recordSavedProject
	_ = initNeutralCwd
	_ = initNonGitBasenameInDir
	_ = resolvePath
	_ = sortedSavedPaths
	_ = whereArgs
	_ = setWhereFlagThenBasename
	_ = setWhereBasenameThenFlag
}
```

