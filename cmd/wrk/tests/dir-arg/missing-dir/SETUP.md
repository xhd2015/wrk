# Scenario

**Feature**: wrk rejects non-existent directory argument

```
# <dir> does not exist on disk
WorkRoot -> wrk /nonexistent -> non-zero exit
```

## Steps

1. Run `wrk` with a path that does not exist.
2. `req.TargetDir` is set to a guaranteed-missing absolute path under `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.TargetDir = filepath.Join(req.WorkRoot, "does-not-exist")
	req.RepoDir = req.WorkRoot
	return nil
}
```