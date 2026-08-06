# Scenario

**Feature**: named create when a live linked WT of source already sits under the same parent as the intended spawn

```
# intended spawn parent = {WorkRoot}/target (existing dir → named subdir create)
# prior linked WTs created as siblings under {WorkRoot}/target (not WRK_HOME)
# reusability / TTY decide create vs skip prompt
myrepo + sibling linked WT(s) under same parent
  -> wrk myrepo {WorkRoot}/target
```

## Preconditions

- Descendants create prior linked worktrees as **siblings** under `{WorkRoot}/target`.
- Spawn target is existing directory `{WorkRoot}/target`.

## Steps

- Prepare existing target parent; leaves add siblings and set TTY/stdin.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureNamedBringReuseHelpersUsed()
	policyBPrepareExistingTargetDir(t, req)
	req.MainRepo = req.TargetDir
	return nil
}
```
