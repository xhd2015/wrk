# Scenario

**Feature**: `--list` + `--exec` is rejected

```
myrepo -> wrk --list --exec true -> non-zero; not valid / mutually exclusive
```

## Steps

1. Initialize git repo.
2. Run `wrk --list --exec true`.

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
	req.RepoDir = mainRepo
	req.Args = []string{"--list", "--exec", "true"}
	return nil
}
```
