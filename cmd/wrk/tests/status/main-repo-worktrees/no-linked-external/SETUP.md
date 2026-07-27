# Scenario

**Feature**: clean main repo with no external linked worktrees

```
# main repo only; no wrk external worktrees under WRK_HOME
myrepo (clean) -> wrk --status from main -> scan block only (unchanged)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main` with one commit.
2. Run `wrk --status` from the main repo root.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initMainRepo(t, mainRepo, "status main root")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	return nil
}
```