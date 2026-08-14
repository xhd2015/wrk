# Scenario

**Feature**: wrk <dir> <target-dir> overrides the worktree spawn location

```
# second positional <target-dir> overrides default {WRK_HOME}/worktrees spawn path
myrepo (main) -> wrk myrepo <target-dir> -> worktree at <target-dir> or <target-dir>/<name>
# <target-dir> resolved relative to shell cwd (process cwd), NOT relative to <dir>
wrk <dir> <target-dir> -> spawn path overridden; WRK_HOME ignored
# create-only: <target-dir> + --list/--done -> wrk: unexpected arguments
# <target-dir> + --bring composes: create at spawn path then bring into that path (create-bring/target-dir)
```

## Preconditions

- Git must be available for all leaves (every leaf spawns or rejects a worktree).
- Source repo `myrepo` lives on branch `main` with one commit; `WRK_DATE=2026-06-30`.

## Steps

- Every leaf initializes the source repo `myrepo` on `main` under `{WorkRoot}`.
- Leaves run `wrk <myrepo> <target-dir> [flags...]` from process cwd `{WorkRoot}` (so a
  relative `<target-dir>` resolves against the shell cwd, not the repo dir).
- `req.TargetDir` = `{WorkRoot}/myrepo` (absolute source repo, first positional).
- `req.SpawnDir` = the `<target-dir>` under test (absolute for most leaves; relative for
  `relative-path/`). `req.Args` carries any trailing flags (`--list`). Create+bring
  lives under `create-bring/success/target-dir/`.
- Expected worktree paths are NOT under `{WRK_HOME}/worktrees`; assert funcs compute them
  directly with `filepath.Join(req.WorkRoot, ...)`.

## Context

- The worktree spawn location changes from `{WRK_HOME}/worktrees/...` to either
  `<target-dir>` (target missing, parent exists) or
  `<target-dir>/{basename}-{token}-{WRK_DATE}[-N]` (target exists).
- Branch naming is always `{token}-{date}[-slug][-N]` via `worktree add -b` (never
  reuse / never `--no-checkout`). Fixed path: if preferred branch exists, suffix the
  **branch only** (`-1`, `-2`, …); path stays fixed. Named subdir under existing
  target: joint path+branch suffix via `candidateBlocked`.
- `WRK_HOME` is ignored when `<target-dir>` is given.
- basename=`myrepo`, token=`main`, date=`2026-06-30` for all leaves.

## New / behavior-change coverage

- `target-missing/parent-exists/basic/` — fixed path, branch free (P0 C4; was leaf `parent-exists/`).
- `target-missing/parent-exists/branch-collision/` — fixed path + pre-existing branch →
  path fixed, branch `main-{date}-1` (P0 C3). Grouping node `parent-exists/` is SETUP-only (no ASSERT.md).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)

	req.TargetDir = repoDir
	req.RepoDir = req.WorkRoot
	return nil
}
```
