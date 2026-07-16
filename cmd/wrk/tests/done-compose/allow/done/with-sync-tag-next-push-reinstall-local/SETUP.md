# Scenario

**Feature**: full post-modifier combo + reinstall-local tail with `--done` passes flag validation

```
# full post subset then reinstall tail; optional -y ignored at flag layer
myrepo -> wrk --done --sync --tag-next --push --reinstall-local -y
  -> not mutually exclusive
  -> not "--push is only valid with --tag-next" false reject
  -> may later fail: not a linked worktree
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run full combo including `--reinstall-local` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done", "--sync", "--tag-next", "--push", "--reinstall-local", "-y"}
	return nil
}
```
