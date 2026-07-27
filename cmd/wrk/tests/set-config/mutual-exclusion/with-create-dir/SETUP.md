# Scenario

**Feature**: set-config rejects create-mode directory positional

```
wrk <dir> --set-config --create --new-terminal -> non-zero
# management must not create a worktree
```

## Steps

1. Init a git repo under WorkRoot for a plausible create dir.
2. Run `wrk <dir> --set-config --create --new-terminal` via TargetDir + Args.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repo)
	req.RepoDir = req.WorkRoot // process cwd outside repo
	req.TargetDir = repo
	req.Args = setConfigArgs("--create", "--new-terminal")
	return nil
}
```
