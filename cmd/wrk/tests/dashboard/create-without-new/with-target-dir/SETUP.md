# Scenario

**Feature**: `wrk <dir>` still creates without `--new`

```
WorkRoot cwd -> wrk myrepo
  -> exit 0; stdout worktree path under WRK_HOME
```

## Steps

1. Init `myrepo` under WorkRoot.
2. Run `wrk <myrepo>` from WorkRoot (TargetDir; no `--new`).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)
	req.MainRepo = repoDir
	req.TargetDir = repoDir
	req.RepoDir = req.WorkRoot
	req.Args = nil
	return nil
}
```
