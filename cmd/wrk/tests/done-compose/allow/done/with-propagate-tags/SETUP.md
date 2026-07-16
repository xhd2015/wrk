# Scenario

**Feature**: `--done --propagate-tags` passes flag validation (primary + propagate compose)

```
# flag layer accepts done + propagate-tags (existing-tags post stage; no tag-next required)
myrepo -> wrk --done --propagate-tags
  -> not mutually exclusive
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --propagate-tags` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--propagate-tags"}
	return nil
}
```
