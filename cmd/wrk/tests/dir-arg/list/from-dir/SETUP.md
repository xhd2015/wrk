# Scenario

**Feature**: wrk --list from main repo via directory argument

```
# single main checkout; wrk invoked from WorkRoot
myrepo (main) -> wrk myrepo --list -> git worktree list
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Run `wrk <myrepo> --list` with process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)

	req.TargetDir = repoDir
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--list"}
	return nil
}
```