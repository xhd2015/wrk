# Scenario

**Feature**: --main --status from an external wrk worktree matches status from main

```
# main + external wt under WRK_HOME; cwd = external
myrepo -> wrk -> external wt
external wt cwd + --main --status -> full main status (scan + append)
```

## Preconditions

- External worktree created via `wrk` (no args) from main (`setupWrkWorktreeFromMain`).
- `WRK_DATE=2026-06-30`; isolated `WRK_HOME`.

## Steps

1. Create main repo + one external wrk worktree.
2. Leaves set Args order; cwd remains the external worktree.
3. Assert stdout == `wrk --status` run from main.

## Context

- Plain `wrk --status` from the external wt is scan-only (no append); the composition
  must produce the multi-block main view instead.

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