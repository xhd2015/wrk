# Scenario

**Feature**: first wrk from fresh repo on main via directory argument

```
# fresh repo on main, wrk invoked from WorkRoot (not repo cwd)
myrepo (main) -> wrk myrepo -> ~/.wrk/worktrees/myrepo-main-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main` with one commit.
2. Run `wrk <myrepo>` with process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)

	req.TargetDir = repoDir
	req.RepoDir = req.WorkRoot
	return nil
}
```