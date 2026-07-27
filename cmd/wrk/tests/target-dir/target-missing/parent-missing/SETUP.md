# Scenario

**Feature**: <target-dir> absent and its parent dir also absent → error

```
# neither {WorkRoot}/missing-parent nor {WorkRoot}/missing-parent/wt exists
myrepo (main) -> wrk myrepo {WorkRoot}/missing-parent/wt -> non-zero, parent does not exist
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Set `req.SpawnDir = {WorkRoot}/missing-parent/wt` (neither segment exists).
3. Run `wrk myrepo {WorkRoot}/missing-parent/wt` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.SpawnDir = filepath.Join(req.WorkRoot, "missing-parent", "wt")
	return nil
}
```
