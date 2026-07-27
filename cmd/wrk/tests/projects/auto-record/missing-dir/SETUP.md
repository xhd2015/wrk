# Scenario

**Feature**: no auto-record when `<dir>` does not exist

```
WorkRoot -> wrk <nonexistent> --list -> error; no projects.json created
```

## Steps

1. Run `wrk <missingPath> --list` from `{WorkRoot}`.

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