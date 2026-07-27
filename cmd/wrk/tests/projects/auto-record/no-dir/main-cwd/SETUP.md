# Scenario

**Feature**: auto-record from main repo cwd

```
myrepo (main cwd) -> wrk --list -> projects.json records myrepo main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --list` with cwd = main repo.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	return nil
}
```