# Scenario

**Feature**: `--merge-back --reinstall-local` is allowed at flag validation

```
# reinstall-local composes with primary --merge-back (parity with --done)
myrepo -> wrk --merge-back --reinstall-local
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --merge-back --reinstall-local` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--merge-back", "--reinstall-local"}
	return nil
}
```
