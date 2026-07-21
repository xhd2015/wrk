# Scenario

**Feature**: named bring when source main already has live linked worktree(s)

```
# any live linked WT of myrepo (e.g. under WRK_HOME/worktrees) triggers Policy B
# default auto-skip; --confirm restores prompt; n creates under target
myrepo + existing linked WT(s)
  -> wrk myrepo <target-dir>
```

## Preconditions

- At least one prior linked worktree of `myrepo` is created by descendants (or this node when shared).
- Spawn target is an existing directory `{WorkRoot}/target` so a non-skip create would land as a named subdir.

## Steps

- This grouping ensures helpers; leaves create prior linked WTs and set TTY/stdin/confirm.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	ensureNamedBringReuseHelpersUsed()
	target := filepath.Join(req.WorkRoot, "target")
	mkdirAll(t, target)
	req.SpawnDir = target
	// Mark grouping: existing-linked branch of Policy B.
	req.MainRepo = req.TargetDir
	return nil
}
```
