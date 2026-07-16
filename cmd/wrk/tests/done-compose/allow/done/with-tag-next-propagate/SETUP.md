# Scenario

**Feature**: `--done --tag-next --propagate-tags` passes flag validation

```
# flag layer accepts full tag-then-propagate post modifiers on primary
myrepo -> wrk --done --tag-next --propagate-tags
  -> not mutually exclusive
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --tag-next --propagate-tags` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--tag-next", "--propagate-tags"}
	return nil
}
```
