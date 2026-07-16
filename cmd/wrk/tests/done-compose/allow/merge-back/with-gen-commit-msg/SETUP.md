# Scenario

**Feature**: `--gen-commit-msg --commit --model=m --merge-back` is allowed at flag validation (P2 pre-stage)

```
# P2: gen-commit pre-stage composes with primary --merge-back when --commit is present
myrepo -> wrk --gen-commit-msg --commit --model=m --merge-back
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree (flag-layer leaf only)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --gen-commit-msg --commit --model=m --merge-back` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--gen-commit-msg", "--commit", "--model=m", "--merge-back"}
	return nil
}
```
