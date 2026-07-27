# Scenario

**Feature**: bare `--push` remains mutually exclusive with `--list`

```
myrepo -> wrk --push --list
  -> non-zero
  -> stderr indicates mutual exclusion / mode conflict
```

## Steps

1. Seed a minimal main repo (with or without origin — flag layer only).
2. Run `wrk --push --list`.

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
	req.Args = []string{"--push", "--list"}
	return nil
}
```
