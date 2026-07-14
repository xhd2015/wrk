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
)

func Setup(t *testing.T, req *Request) error {
	req.TargetDir = filepath.Join(req.WorkRoot, "does-not-exist")
	req.RepoDir = req.WorkRoot
	return nil
}
```