# Scenario

**Feature**: `--status` + `--exec` is rejected

```
myrepo -> wrk --status --exec true -> non-zero; not valid / mutually exclusive
```

## Steps

1. Initialize git repo.
2. Run `wrk --status --exec true`.

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
	req.Args = []string{"--status", "--exec", "true"}
	return nil
}
```
