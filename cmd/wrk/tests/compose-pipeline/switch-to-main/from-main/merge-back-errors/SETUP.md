# Scenario

**Feature**: bare `--merge-back` on main hard-errors (linked worktree required)

```
myrepo (main) -> wrk --merge-back
  -> non-zero
  -> stderr names --merge-back and linked worktree requirement
```

## Steps

1. Main repo.
2. Run `wrk --merge-back` from main.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--merge-back"}
	return nil
}
```
