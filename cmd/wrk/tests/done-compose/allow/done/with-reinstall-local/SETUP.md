# Scenario

**Feature**: `--done --reinstall-local` is allowed at flag validation (reinstall post-success tail)

```
# P1: reinstall-local composes as post-success tail after primary --done
myrepo -> wrk --done --reinstall-local
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree (flag-layer leaf only)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --reinstall-local` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--reinstall-local"}
	return nil
}
```
