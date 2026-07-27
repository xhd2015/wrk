# Scenario

**Feature**: `--merge-back --propagate-tags` passes flag validation

```
myrepo -> wrk --merge-back --propagate-tags
  -> not mutually exclusive
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --merge-back --propagate-tags` from the main repo.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--merge-back", "--propagate-tags"}
	return nil
}
```
