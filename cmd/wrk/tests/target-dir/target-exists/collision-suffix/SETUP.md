# Scenario

**Feature**: <target-dir> exists and the default sub-dir name collides → -N suffix

```
# {WorkRoot}/target exists; pre-create {WorkRoot}/target/myrepo-main-2026-06-30 to block suffix 0
myrepo (main) -> wrk myrepo {WorkRoot}/target -> {WorkRoot}/target/myrepo-main-2026-06-30-1
# both path and branch get the -1 suffix
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Pre-create `{WorkRoot}/target` as a directory.
3. Pre-create the colliding sub-dir `{WorkRoot}/target/myrepo-main-2026-06-30` (empty).
4. Set `req.SpawnDir = {WorkRoot}/target`.
5. Run `wrk myrepo {WorkRoot}/target` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	target := filepath.Join(req.WorkRoot, "target")
	mkdirAll(t, target)
	// pre-create the suffix-0 sub-dir name to force collision -> -1
	mkdirAll(t, filepath.Join(target, "myrepo-main-"+wrkDate))
	req.SpawnDir = target
	return nil
}
```
