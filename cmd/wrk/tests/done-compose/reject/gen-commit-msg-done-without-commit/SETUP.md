# Scenario

**Feature**: `--gen-commit-msg --done` without `--commit` is rejected (P2 requires --commit with primary)

```
myrepo -> wrk --gen-commit-msg --done
  -> non-zero
  -> stderr requires --commit with primary / --done
  -> must not silently accept into early gen-commit-only path that only mutexes
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --gen-commit-msg --done` (no `--commit`) from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--gen-commit-msg", "--done"}
	return nil
}
```
