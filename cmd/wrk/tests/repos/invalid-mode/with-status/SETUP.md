# Scenario

**Feature**: wrk --repos rejects --status in the same invocation

```
wrk --repos --status -> error (mutually exclusive)
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Run `wrk --repos --status` from the repo root.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "myrepo")
	reposInitRepo(t, repo)
	req.RepoDir = repo
	req.Args = []string{"--repos", "--status"}
	return nil
}
```
