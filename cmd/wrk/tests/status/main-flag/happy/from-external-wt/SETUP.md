# Scenario

**Feature**: --main --status from an external wrk worktree statuses main with inv-cwd Dirs

```
# main + external wt under WRK_HOME; cwd = external
myrepo -> wrk -> external wt
external wt cwd + --main --status -> main content; Dir labels vs external cwd
```

## Preconditions

- External worktree created via `wrk` (no args) from main (`setupWrkWorktreeFromMain`).
- `WRK_DATE=2026-06-30`; isolated `WRK_HOME`.

## Steps

1. Create main repo + one external wrk worktree.
2. Leaves set Args order; cwd remains the external worktree.
3. Assert content matches main status; Dir lines rewritten for inv cwd.

## Context

- Plain `wrk --status` from the external wt is scan-only (no append); the composition
  must produce the multi-block main view instead.
- Stdout is **not** required to be byte-equal to `(cd main && wrk --status)`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, branch := setupExternalMainFlagFixture(t, req)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch
	req.RepoDir = wtDir
	return nil
}
```
