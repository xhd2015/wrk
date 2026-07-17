# Scenario

**Feature**: bare `--gen-commit-msg --sync` is allowed multi-stage compose (activeRoot stays cwd; no done required)

```
# Target model: stages may compose without --done/--merge-back
myrepo -> wrk --gen-commit-msg --sync
  -> must NOT stderr "mutually exclusive"
  -> may later fail for missing --commit / no staged changes (not a mode mutex)
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
