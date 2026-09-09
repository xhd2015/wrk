# Scenario

**Feature**: wrk --status reports nested git checkouts under a plain directory

```
# plain workspace container with sibling independent repos
plain/ + plain/alpha + plain/beta -> wrk --status
  -> blocks for alpha then beta (path-sorted); no Remote; no external header
```

## Steps

1. Create `{WorkRoot}/plain` as a non-git directory.
2. Initialize `{WorkRoot}/plain/alpha` and `{WorkRoot}/plain/beta` as independent
   git repos on branch `main`.
3. Run `wrk --status` from `{WorkRoot}/plain`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plain := filepath.Join(req.WorkRoot, "plain")
	alpha := filepath.Join(plain, "alpha")
	beta := filepath.Join(plain, "beta")

	mkdirAll(t, plain)
	statusInitRepoWithSubject(t, alpha, "alpha status repo")
	statusInitRepoWithSubject(t, beta, "beta status repo")

	req.RepoDir = plain
	req.MainRepo = alpha
	req.DepPath = beta
	return nil
}
```
