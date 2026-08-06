# Scenario

**Feature**: named create when the target path already exists as a linked worktree → error, no create

```
# worktree-A is already a live linked WT of myrepo (exact path occupied)
# wrk myrepo {WorkRoot}/target/worktree-A -> non-zero; already exists / not a free target
# must not nest a new named subdir under the existing worktree
myrepo (main) + existing linked WT at worktree-A
  -> wrk myrepo {WorkRoot}/target/worktree-A
  -> error; no nested create
```

## Steps

1. Pre-create parent `{WorkRoot}/target` and a live linked worktree at `{WorkRoot}/target/worktree-A`.
2. Set `req.SpawnDir` to that exact existing worktree path.
3. Run `wrk myrepo <exact-path>` in-process.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureNamedBringReuseHelpersUsed()
	parent := policyBTargetParent(req)
	existsPath := filepath.Join(parent, "worktree-A")
	// Occupy the exact path as a linked worktree of source (not a plain empty container dir).
	policyBAddSiblingLinked(t, req, existsPath, "policy-b-exact-exists")
	req.SpawnDir = existsPath
	req.WtDir = existsPath
	return nil
}
```
