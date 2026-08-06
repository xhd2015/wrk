# Scenario

**Feature**: named create with no prior linked worktree of source creates as today (no Policy B)

```
# no live linked WT of myrepo -> wrk myrepo <target> creates new WT under target
# no Policy B banner: no "would reuse" / "skip creating"
myrepo (main only) -> wrk myrepo {WorkRoot}/target -> {WorkRoot}/target/myrepo-main-{date}
```

## Steps

1. Ensure source `myrepo` has no linked worktrees (parent init only).
2. Pre-create `{WorkRoot}/target` as an existing directory.
3. Set `req.SpawnDir = {WorkRoot}/target`.
4. Run `wrk myrepo <target>` via `Run`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureNamedBringReuseHelpersUsed()
	policyBPrepareExistingTargetDir(t, req)
	_ = filepath.Join
	return nil
}
```
