# Scenario

**Feature**: wrk --list from main repo with one linked worktree

```
# main repo + linked worktree
myrepo + linked-wt -> wrk --list -> lists both paths
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Add linked worktree at `{WorkRoot}/linked-wt` on branch `linked-side`.
3. Run `wrk --list` with cwd set to the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)

	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "linked-side", linkedWT)

	req.RepoDir = mainRepo
	req.Args = []string{"--list"}
	return nil
}
```