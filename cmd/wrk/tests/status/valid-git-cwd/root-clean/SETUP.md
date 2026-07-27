# Scenario

**Feature**: wrk --status shows the root checkout as dot when run from the root

```
# process cwd is the checkout top
myrepo (clean) -> wrk --status -> Dir "."
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `initial status root`.
3. Run `wrk --status` from the repo root.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "initial status root")

	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
