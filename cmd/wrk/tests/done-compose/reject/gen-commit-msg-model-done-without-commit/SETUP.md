# Scenario

**Feature**: `--gen-commit-msg --model=m --done` without `--commit` is rejected

```
# model alone does not satisfy the --commit requirement with primary
myrepo -> wrk --gen-commit-msg --model=m --done
  -> non-zero
  -> requires --commit with primary
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --gen-commit-msg --model=m --done` (no `--commit`) from the main repo.

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
	req.Args = []string{"--gen-commit-msg", "--model=m", "--done"}
	return nil
}
```
