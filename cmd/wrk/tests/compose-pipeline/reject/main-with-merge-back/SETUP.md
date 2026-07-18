# Scenario

**Feature**: `--main` and `--merge-back` remain mutually exclusive

```
myrepo -> wrk --main --merge-back
  -> non-zero
  -> stderr mutual exclusion (or not valid) naming both flags
```

## Steps

1. Minimal main repo.
2. Run both flags together.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "--merge-back"}
	return nil
}
```
