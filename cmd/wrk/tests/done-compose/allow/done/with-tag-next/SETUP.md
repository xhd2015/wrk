# Scenario

**Feature**: `--done --tag-next` is allowed at flag validation (not mutually exclusive)

```
# main repo (not linked wt) so mode clash used to fire first
myrepo -> wrk --done --tag-next
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree (proves flag layer passed)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --tag-next` from the main repo.

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
	req.Args = []string{"--done", "--tag-next"}
	return nil
}
```
