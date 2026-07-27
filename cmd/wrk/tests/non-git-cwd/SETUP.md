# Scenario

**Feature**: wrk rejects non-git cwd for create

```
# plain directory without .git
plain cwd -> wrk --new -> error (not a git repository)
```

## Steps

1. Create a non-git directory under `{WorkRoot}/plain`.
2. Run `wrk --new` with cwd set to that directory (create entry; bare no-args is dashboard).

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
	req.Args = []string{"--new"}
	return nil
}
```
