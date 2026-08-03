# Scenario

**Feature**: --commit -m --done --dry-run accepted at flag layer

```
myrepo -> wrk --commit -m "feat: compose" --done --dry-run
  -> must NOT stderr "mutually exclusive"
  -> must NOT stderr "--dry-run is only valid with …" for this host combo
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run compose dry-run with manual message.

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
	req.Args = []string{"--commit", "-m", "feat: compose", "--done", "--dry-run"}
	return nil
}
```
