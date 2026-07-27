# Scenario

**Feature**: `--main` is not valid with `--gen-commit-msg` (named reject preferred)

```
workspace/ -> wrk --main --gen-commit-msg --dry-run
  -> non-zero
  -> preferred: wrk: --main is not valid with --gen-commit-msg
  -> (generic mutually exclusive also acceptable until named error lands)
```

## Steps

1. Minimal main repo so rejection is about flag matrix, not missing git.
2. Run `--main` with `--gen-commit-msg`.

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
	req.Args = []string{"--main", "--gen-commit-msg", "--dry-run"}
	return nil
}
```
