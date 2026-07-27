# Scenario

**Feature**: full post-modifier combo with `--done` passes flag validation

```
# at least one multi-modifier combo; flag order free (this leaf: done first)
myrepo -> wrk --done --sync --tag-next --push
  -> not mutually exclusive; push not tag-next-only
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --sync --tag-next --push` from the main repo.

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
	req.Args = []string{"--done", "--sync", "--tag-next", "--push"}
	return nil
}
```
