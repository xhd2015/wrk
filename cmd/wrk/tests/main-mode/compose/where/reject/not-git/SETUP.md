# Scenario

**Feature**: wrk --main --where rejects non-git cwd (same messaging as bare --main)

```
plain cwd -> wrk --main --where -> non-zero; not a git repository
```

## Steps

1. Create non-git directory under WorkRoot.
2. Run `wrk --main --where` from that directory.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plainDir := filepath.Join(req.WorkRoot, "plain")
	mkdirAll(t, plainDir)
	req.RepoDir = plainDir
	setMainWhereArgs(req, "--main", "--where")
	return nil
}
```
