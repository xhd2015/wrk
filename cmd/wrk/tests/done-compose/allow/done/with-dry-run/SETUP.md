# Scenario

**Feature**: `--done --dry-run` is accepted at flag layer (not “only valid with …”)

```
# flag matrix: composition dry-run must not be rejected as invalid dry-run host mode
# full multi-stage plan stdout is asserted under done-pipeline/dry-run/
myrepo -> wrk --done --dry-run
  -> must NOT stderr "--dry-run is only valid with …"
  -> may later fail: not a linked worktree (flag-layer leaf only)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --dry-run` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
