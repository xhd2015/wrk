# Scenario

**Feature**: named create `wrk <dir> <target-dir>` Policy B — scoped same-parent reusable sibling reuse

```
# Policy B (named create only): scan live linked WTs of source mainRepo that are
# **direct siblings under parent(intendedSpawnPath)** and reusable (clean + HEAD==source).
# spawn path already exists -> error (no create)
# no reusable same-parent sibling -> create as today; no Policy B banner/prompt
#   (includes: no prior WT; WT only under other parent e.g. WRK_HOME; dirty; clean-but-differs)
# >=1 reusable sibling + TTY -> wrk: warning: would reuse <path>; skip creating? [Y/n]
#   Y/y/empty -> skip; stdout = lex-smallest reusable path
#   n/N -> create as today
#   multi -> primary = lex-smallest; also present: other reusable paths
# >=1 reusable sibling + non-TTY -> create (no refuse; automation-friendly)
myrepo (main) [+ optional sibling linked WTs under same parent as spawn]
  -> wrk myrepo <target-dir>
```

## Preconditions

- Parent `target-dir` setup initializes source repo `myrepo` on `main` and sets `req.TargetDir`.
- Leaves set `req.SpawnDir` and optionally pre-create linked worktrees as **siblings under the same parent as the intended spawn path** (not only under `WRK_HOME`).
- Process cwd is `{WorkRoot}` so relative spawn paths resolve against shell cwd.

## Steps

- Grouping only: descendants narrow spawn collision / sibling location / reusability / TTY.

## Context

- **Same parent**: cleaned absolute `filepath.Dir` of the **intended spawn path** only.
  Existing target dir → intended spawn = first free `{basename}-{token}-{date}[-N]` under it;
  missing target with parent exists → intended spawn = exact `<target-dir>`.
- **Reusable sibling**: live linked WT `P` of same `mainRepo` where `parent(P)==parent(spawn)`,
  porcelain clean, and `HEAD(P)==HEAD(source)`.
- Dirty or clean-but-differs-from-source → **not** reusable → create, no banner.
- Worktrees under other parents (`WRK_HOME/worktrees`, other workspace folders) do **not** trigger Policy B.
- stdout: path only. Prompt + warnings + errors on stderr.
- No override flags (`-y` / `--force` out of scope for this feature).

```go
import (
	"os"
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

// policyBTargetParent returns {WorkRoot}/target — common parent for existing-dir spawn
// (intended spawn = target/myrepo-main-{date}[-N]).
func policyBTargetParent(req *Request) string {
	return filepath.Join(req.WorkRoot, "target")
}

// policyBPrepareExistingTargetDir creates {WorkRoot}/target and sets req.SpawnDir so a
// non-skip create lands as a named subdir under it.
func policyBPrepareExistingTargetDir(t *testing.T, req *Request) string {
	t.Helper()
	parent := policyBTargetParent(req)
	mkdirAll(t, parent)
	req.SpawnDir = parent
	return parent
}

// policyBAddSiblingLinked creates a live linked worktree of myrepo at absPath with a
// new branch. Parent of absPath must exist (or be created). Branch is chosen by the
// caller so preferred create branch main-{date} can stay free when useful.
// Returns cleaned absPath.
func policyBAddSiblingLinked(t *testing.T, req *Request, absPath, branch string) string {
	t.Helper()
	parent := filepath.Dir(absPath)
	mkdirAll(t, parent)
	if _, err := os.Stat(absPath); err == nil {
		t.Fatalf("sibling path already exists: %s", absPath)
	}
	runGitIsolated(t, req.TargetDir, "worktree", "add", "-b", branch, absPath)
	assertGitFileIsWorktreeLink(t, absPath)
	assertWorktreeListContains(t, req.TargetDir, absPath)
	return absPath
}

// policyBAddSiblingUnderParent adds a linked WT at {parent}/{name} with branch.
func policyBAddSiblingUnderParent(t *testing.T, req *Request, parent, name, branch string) string {
	t.Helper()
	return policyBAddSiblingLinked(t, req, filepath.Join(parent, name), branch)
}

// namedBringExistingWorktrees creates N sequential wrk worktrees of myrepo under WRK_HOME
// (other parent relative to {WorkRoot}/target spawn). Returns abs paths in create order.
// Used only for other-parent fixtures — these must NOT trigger scoped Policy B alone.
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

// assertNoPolicyBBanner fails if stderr/stdout contain Policy B skip/reuse prompt tokens.
func assertNoPolicyBBanner(t *testing.T, combined string) {
	t.Helper()
	assertNotContains(t, combined, "skip creating")
	assertNotContains(t, combined, "would reuse")
	assertNotContains(t, combined, "already has a linked worktree")
	assertNotContains(t, combined, "refusing non-interactive")
}

func ensureNamedBringReuseHelpersUsed() {
	_ = namedBringExistingWorktrees
	_ = policyBPrepareExistingTargetDir
	_ = policyBAddSiblingUnderParent
	_ = policyBAddSiblingLinked
	_ = policyBTargetParent
	_ = assertNoPolicyBBanner
	_ = filepath.Join
}
```
