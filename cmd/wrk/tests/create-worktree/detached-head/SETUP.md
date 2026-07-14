# Scenario

**Feature**: wrk from detached HEAD uses 7-char commit hash token

```
# cwd on detached HEAD at myrepo
myrepo (detached HEAD) -> wrk -> myrepo-{short-hash}-2026-06-30
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Detach HEAD with `git checkout --detach`.
3. Record 7-char short hash via `git rev-parse --short=7 HEAD` into `req.HashToken`.
4. Run `wrk` from `myrepo`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)
	runGitIsolated(t, repoDir, "checkout", "--detach")
	req.RepoDir = repoDir
	req.HashToken = gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	return nil
}
```