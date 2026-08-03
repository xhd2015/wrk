# Scenario

**Feature**: --commit -m "x" --done is allowed at flag validation

```
myrepo -> wrk --commit -m "feat: compose" --done
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree (flag-layer leaf only)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --commit -m "feat: compose" --done` from the main repo.

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
	req.Args = []string{"--commit", "-m", "feat: compose", "--done"}
	return nil
}
```
