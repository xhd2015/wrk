# Scenario

**Feature**: `--done --json` is rejected

```
myrepo -> wrk --done --json
  -> non-zero
  -> stderr names --json and --done (e.g. not valid with --done)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --json` from the main repo.

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
	req.Args = []string{"--done", "--json"}
	return nil
}
```
