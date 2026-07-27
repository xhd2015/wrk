# Scenario

**Feature**: bare `--reinstall-local --sync` is allowed multi-stage compose (no done required)

```
# Target model: reinstall and sync compose under fixed stage order without primary
myrepo -> wrk --reinstall-local --sync
  -> must NOT stderr "mutually exclusive"
  -> may later fail for empty reinstall plan / sync conditions
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --reinstall-local --sync` from the main repo.

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
	req.Args = []string{"--reinstall-local", "--sync"}
	return nil
}
```
