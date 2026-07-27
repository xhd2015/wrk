# Scenario

**Feature**: gen-commit pre-stage + full post-modifier combo with `--done` passes flag validation

```
# P2 pre + posts: gen-commit then primary then sync/tag-next/push
myrepo -> wrk --gen-commit-msg --commit --model=m --done --sync --tag-next --push
  -> not mutually exclusive
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run gen-commit + done + post modifiers from the main repo.

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
	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--done", "--sync", "--tag-next", "--push",
	}
	return nil
}
```
