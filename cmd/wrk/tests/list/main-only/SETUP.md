# Scenario

**Feature**: wrk --list from main repo with no linked worktrees

```
# single main checkout only
myrepo (main) -> wrk --list -> git worktree list
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Run `wrk --list` with cwd set to the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)

	req.RepoDir = mainRepo
	req.Args = []string{"--list"}
	return nil
}
```