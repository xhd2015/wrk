# Scenario

**Feature**: <target-dir> absent but parent exists → spawn worktree exactly at <target-dir>

```
# parent {WorkRoot} exists, <target-dir> {WorkRoot}/wt does not
myrepo (main) -> wrk myrepo {WorkRoot}/wt -> worktree at {WorkRoot}/wt (no naming suffix on path)
# branch: main-{date}[-N] via always-new -b; WRK_HOME ignored
```

## Children

- `basic/` — preferred branch free (P0 C4).
- `branch-collision/` — preferred branch pre-exists → branch `-1` only (P0 C3).

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Set `req.SpawnDir = {WorkRoot}/wt` (does not exist; parent `{WorkRoot}` exists).
3. Leaves may pre-create branch refs before run.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	return nil
}
```
