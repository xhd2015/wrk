# Scenario

**Feature**: wrk --main --status when already at main root equals plain --status

```
main repo cwd -> wrk --main --status
  == wrk --status at main
```

## Steps

1. Create main repo (go.mod seed via `setupWrkWorktreeFromMain` not required; plain main is enough).
2. cwd = main root; Args = `--main`, `--status`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, mainRepo, "main flag already at main")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	setMainStatusArgs(req, "--main", "--status")
	return nil
}
```