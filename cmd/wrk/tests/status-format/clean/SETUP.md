# Scenario

**Feature**: wrk --status prints `Status: clean` for an empty porcelain checkout

```
# committed checkout, no dirty files
myrepo (clean) -> wrk --status -> Status: clean
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Commit `README.md` with subject `status format clean base`.
3. Run `wrk --status` from the repo root (clean porcelain).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusFormatInitRepo(t, repo, "status format clean base")
	req.RepoDir = repo
	req.MainRepo = repo
	return nil
}
```
