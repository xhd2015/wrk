# Scenario

**Feature**: main repo without upstream shows Remote: (no upstream)

```
plain main repo (no tracking remote) -> wrk --status -> root Remote: (no upstream)
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "no-upstream")
	initFetchVerboseRepo(t, repo, "no upstream base")
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--status"}
	return nil
}
```