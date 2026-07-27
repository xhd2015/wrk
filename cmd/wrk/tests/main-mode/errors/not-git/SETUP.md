# Scenario

**Feature**: wrk --main rejects cwd that is not a git repository

```
# plain directory without .git
plain cwd -> wrk --main -> non-zero; not a git repository
```

## Steps

1. Create a non-git directory under `{WorkRoot}/plain`.
2. Run `wrk --main` with cwd set to that directory.

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
	setMainArgs(req)
	return nil
}
```
