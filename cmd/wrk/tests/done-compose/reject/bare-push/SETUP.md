# Scenario

**Feature**: bare `wrk --push` is still rejected

```
myrepo -> wrk --push
  -> non-zero
  -> stderr explains --push is not standalone
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --push` alone from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--push"}
	return nil
}
```
