# Scenario

**Feature**: `--tag-next --list` remains mutually exclusive (regression)

```
myrepo -> wrk --tag-next --list
  -> non-zero
  -> stderr mutually exclusive
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --tag-next --list` from the main repo.

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
	req.Args = []string{"--tag-next", "--list"}
	return nil
}
```
