# Scenario

**Feature**: `--json` is only valid with bare `--tag-next`; multi-stage + `--json` is rejected

```
myrepo -> wrk --sync --tag-next --json
  -> non-zero
  -> stderr names --json and rejects multi-stage combination
```

## Steps

1. Minimal main repo.
2. Run multi-stage with `--json`.

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
	req.Args = []string{"--sync", "--tag-next", "--json"}
	return nil
}
```
