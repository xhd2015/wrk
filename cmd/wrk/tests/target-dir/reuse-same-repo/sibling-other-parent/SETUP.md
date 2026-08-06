# Scenario

**Feature**: live linked WT of source only under another parent (WRK_HOME) does not trigger Policy B

```
# prior WT only under WRK_HOME/worktrees (parent != parent(intended spawn))
# wrk myrepo {WorkRoot}/target -> create as today; no would-reuse / skip-creating banner
myrepo + WRK_HOME linked WT
  -> wrk myrepo {WorkRoot}/target
  -> new {WorkRoot}/target/myrepo-main-{date}[-N]; no Policy B prompt
```

## Steps

1. Pre-create one linked worktree of `myrepo` under `WRK_HOME/worktrees` (other parent).
2. Pre-create existing target dir `{WorkRoot}/target` and set `SpawnDir`.
3. Run named create in-process (non-TTY).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureNamedBringReuseHelpersUsed()
	paths := namedBringExistingWorktrees(t, req, 1)
	req.WtDir = paths[0] // other-parent path (must not be reused / must not banner)
	policyBPrepareExistingTargetDir(t, req)
	return nil
}
```
