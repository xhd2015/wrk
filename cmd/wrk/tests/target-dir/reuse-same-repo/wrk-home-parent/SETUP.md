# Scenario

**Feature**: named create whose intended spawn parent is this process's `{WRK_HOME}/worktrees` dump skips Policy B

```
# parent(intendedSpawn) == {WRK_HOME}/worktrees → no sibling scan / no would-reuse prompt
# missing exact name under dump → create at that path (same freedom as bare wrk)
# occupied dump path → already exists (skip Policy B ≠ reuse this named WT)
myrepo + reusable dump sibling
  -> wrk myrepo {WRK_HOME}/worktrees/<name>
```

## Preconditions

- Intended spawn parent is `{req.WrkHome}/worktrees` (this process's dump, not a hardcoded `~/.wrk`).
- Grouping pre-creates one porcelain-clean, HEAD==source linked WT of `myrepo` under that dump
  (preferred create branch `main-{date}` is taken by that sibling).

## Steps

- Create one dump sibling via `namedBringExistingWorktrees`. Leaves set `SpawnDir` to either
  a new missing name under the dump or the occupied sibling path.

## Context

- Skip Policy B when `abs(clean(parent(intendedSpawn))) == abs(clean({WRK_HOME}/worktrees))`.
- Create proceeds as today: missing target + parent exists → exact path.
- Occupied intended spawn still errors `already exists`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureNamedBringReuseHelpersUsed()
	mkdirAll(t, policyBDumpParent(req))
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0]
	req.MainRepo = req.TargetDir
	return nil
}
```
