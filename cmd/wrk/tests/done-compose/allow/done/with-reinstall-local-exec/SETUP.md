# Scenario

**Feature**: `--done --reinstall-local --exec …` is accepted at flag layer (reinstall before exec/land)

```
# locked order after primary success: post → reinstall-local → --exec → land
# flag layer must not reject exec when primary + reinstall compose
myrepo -> wrk --done --reinstall-local --exec true
  -> must NOT stderr "mutually exclusive"
  -> must NOT stderr "--exec is not valid with --reinstall-local"
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --done --reinstall-local --exec true` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--reinstall-local", "--exec", "true"}
	return nil
}
```
