# Scenario

**Feature**: <target-dir> exists but is a regular file → error (not a directory)

```
# {WorkRoot}/target is a file, not a directory
myrepo (main) -> wrk myrepo {WorkRoot}/target -> non-zero, not a directory / cannot spawn
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Pre-create `{WorkRoot}/target` as a regular file (not a directory).
3. Set `req.SpawnDir = {WorkRoot}/target`.
4. Run `wrk myrepo {WorkRoot}/target` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	target := filepath.Join(req.WorkRoot, "target")
	writeFile(t, target, "i-am-a-file")
	req.SpawnDir = target
	return nil
}
```
