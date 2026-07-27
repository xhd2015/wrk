# Scenario

**Feature**: `--main` and `--done` remain mutually exclusive

```
myrepo -> wrk --main --done
  -> non-zero
  -> stderr mutual exclusion (or not valid) naming both flags
```

## Steps

1. Minimal main repo.
2. Run both flags together.

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
	req.Args = []string{"--main", "--done"}
	return nil
}
```
