# Scenario

**Feature**: bare `--gen-commit-msg --sync` remains mutually exclusive (no primary)

```
# P2 must not open bare gen-commit + sync; sync is post-only with primary
myrepo -> wrk --gen-commit-msg --sync
  -> non-zero
  -> stderr mutually exclusive (or equivalent mode conflict)
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --gen-commit-msg --sync` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--gen-commit-msg", "--sync"}
	return nil
}
```
