# Scenario

**Feature**: `--gen-commit-msg --commit --model=m --done` is allowed at flag validation (P2 pre-stage)

```
# P2: gen-commit pre-stage composes with primary --done when --commit is present
myrepo -> wrk --gen-commit-msg --commit --model=m --done
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree / no staged changes (flag-layer leaf only)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --gen-commit-msg --commit --model=m --done` from the main repo.

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
	req.Args = []string{"--gen-commit-msg", "--commit", "--model=m", "--done"}
	return nil
}
```
