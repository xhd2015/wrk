# Scenario

**Feature**: <target-dir> exists as a dir → spawn default-named sub-dir under it

```
# {WorkRoot}/target exists (empty dir)
myrepo (main) -> wrk myrepo {WorkRoot}/target -> {WorkRoot}/target/myrepo-main-2026-06-30
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Pre-create `{WorkRoot}/target` as an empty directory.
3. Set `req.SpawnDir = {WorkRoot}/target`.
4. Run `wrk myrepo {WorkRoot}/target` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	target := filepath.Join(req.WorkRoot, "target")
	mkdirAll(t, target)
	req.SpawnDir = target
	return nil
}
```
