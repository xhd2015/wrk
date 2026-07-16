# Scenario

**Feature**: bare `--reinstall-local --sync` remains mutually exclusive (no primary)

```
# composition with primary must not open bare reinstall + sync
myrepo -> wrk --reinstall-local --sync
  -> non-zero
  -> stderr mutually exclusive
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --reinstall-local --sync` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--reinstall-local", "--sync"}
	return nil
}
```
