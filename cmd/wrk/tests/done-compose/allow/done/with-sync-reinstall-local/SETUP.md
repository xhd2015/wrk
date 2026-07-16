# Scenario

**Feature**: `--done --sync --reinstall-local` passes flag validation

```
# primary + post modifier + reinstall tail; flag order free
myrepo -> wrk --done --sync --reinstall-local
  -> not mutually exclusive
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --sync --reinstall-local` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--sync", "--reinstall-local"}
	return nil
}
```
