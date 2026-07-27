# Scenario

**Feature**: cwd file named basename triggers guided error instead of git-repo failure

```
# ./<basename> is a regular file (not directory) -> skip cwd-path resolution
wrk <basename> -> stat(cwd/<basename>) is file -> lookup projects.json by basename

# match count drives stderr shape (no worktree, no further git error)
1 match   -> multi-line stderr: file path, one project line, concrete-path hint
2+ matches -> multi-line stderr: file path, all project lines, <full-path> hint
0 matches -> single-line stderr: file path only

# directory blocking unchanged
./<basename> is a directory (even non-git) -> use cwd path, no lookup
```

## Preconditions

- Inherits helpers from parent `projects/basename-fallback/SETUP.md` (`initSavedGitRepo`, `recordSavedProject`, `initNeutralCwd`, `sortedSavedPaths`).
- `resolvePath` comes from `projects/SETUP.md`.

## Steps

- Descendants create a regular file `./<basename>` in cwd via `initBasenameFile`.
- Configure saved projects and CLI args (`TargetDir`, `TaskDesc`/`TaskFlag`, `Args`) per leaf.

## Context

- Guided errors exit non-zero with empty stdout; no worktree is created.
- Hint reconstructs user flags/args (`-t`, `--status`, etc.) from the invocation.
- Ambiguous hints use literal `<full-path>` placeholder (no default pick).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureCwdFileExistsHelpersUsed()
	return nil
}

// initBasenameFile creates a regular file (not a directory) at dir/basename.
func initBasenameFile(t *testing.T, dir, basename, content string) string {
	t.Helper()
	if content == "" {
		content = "stub file for basename collision test\n"
	}
	path := filepath.Join(dir, basename)
	writeFile(t, path, content)
	return path
}

func ensureCwdFileExistsHelpersUsed() {
	_ = initBasenameFile
}
```