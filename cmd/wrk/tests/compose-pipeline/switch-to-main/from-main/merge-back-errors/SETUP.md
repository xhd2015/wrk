# Scenario

**Feature**: bare `--merge-back` on main hard-errors (linked worktree required)

```
myrepo (main) -> wrk --merge-back
  -> non-zero
  -> stderr names --merge-back and linked worktree requirement
```

## Steps

1. Main repo.
2. Run `wrk --merge-back` from main.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--merge-back"}
	return nil
}
```
