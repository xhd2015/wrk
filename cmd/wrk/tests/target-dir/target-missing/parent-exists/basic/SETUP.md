# Scenario

**Feature**: <target-dir> absent but parent exists → spawn worktree exactly at <target-dir> (P0 C4)

```
# parent {WorkRoot} exists, <target-dir> {WorkRoot}/wt does not
myrepo (main) -> wrk myrepo {WorkRoot}/wt -> worktree at {WorkRoot}/wt (no naming suffix on path)
# branch still defaults to main-{date} via -b; WRK_HOME ignored
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the grandparent setup.
2. Set `req.SpawnDir = {WorkRoot}/wt` (parent also sets this; leaf re-asserts free-branch case).
3. Run `wrk myrepo {WorkRoot}/wt` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// Preferred branch is free (no pre-created refs). Fixed path at {WorkRoot}/wt.
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	return nil
}
```
