# Scenario

**Feature**: named bring `wrk <dir> <target-dir>` avoids duplicate linked worktrees of the same source main repo

```
# Policy B: any live linked worktree of source mainRepo (not only under target / external/)
# none -> create as today
# some + TTY -> prompt "wrk: warning: <basename> already has a linked worktree at <path>, skip creating another? [Y/n]"
#   Y/y/empty -> skip create; stdout = lex-smallest existing path
#   n/N -> create as today under/at target-dir
# some + non-TTY -> hard error; empty stdout; no new WT
# multi -> primary path in prompt is lex-smallest; optional also-present warnings
myrepo (main) [+ existing linked WTs]
  -> wrk myrepo <target-dir>
```

## Preconditions

- Parent `target-dir` setup initializes source repo `myrepo` on `main` and sets `req.TargetDir`.
- Leaves set `req.SpawnDir` and optionally pre-create linked worktrees of `myrepo`.
- Process cwd is `{WorkRoot}` so relative spawn paths resolve against shell cwd.

## Steps

- Grouping only: descendants narrow prior-linked / TTY / user answer.

## Context

- Identity: cleaned absolute `ResolveMainRepo` of source (`myrepo`).
- stdout: path only. Prompt + warnings + errors on stderr.
- No override flags (`-y` / `--force` out of scope for this feature).

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Parent target-dir Setup already inits myrepo + TargetDir/RepoDir.
	ensureNamedBringReuseHelpersUsed()
	return nil
}

// namedBringExistingWorktrees creates N sequential wrk worktrees of myrepo under WRK_HOME
// and returns their absolute paths in create order (suffix 0..N-1).
func namedBringExistingWorktrees(t *testing.T, req *Request, n int) []string {
	t.Helper()
	var paths []string
	for i := 0; i < n; i++ {
		p := runWrkWithArgs(t, req, req.TargetDir)
		want := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, i)
		if p != want {
			t.Fatalf("pre-create worktree %d: expected %q, got %q", i, want, p)
		}
		paths = append(paths, p)
	}
	return paths
}

func ensureNamedBringReuseHelpersUsed() {
	_ = namedBringExistingWorktrees
	_ = filepath.Join
}
```
