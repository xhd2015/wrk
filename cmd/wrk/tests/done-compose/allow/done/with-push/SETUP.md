# Scenario

**Feature**: `--done --push` is allowed at flag validation (push no longer tag-next-only)

```
# --push may compose with primary without requiring --tag-next
myrepo -> wrk --done --push
  -> must NOT stderr "--push is only valid with --tag-next"
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --push` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--push"}
	return nil
}
```
