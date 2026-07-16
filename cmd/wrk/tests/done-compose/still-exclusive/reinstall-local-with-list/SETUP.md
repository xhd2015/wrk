# Scenario

**Feature**: bare `--reinstall-local --list` remains mutually exclusive (regression)

```
# keep list exclusivity; composition must not open this pair
myrepo -> wrk --reinstall-local --list
  -> non-zero
  -> stderr mutually exclusive
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --reinstall-local --list` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--reinstall-local", "--list"}
	return nil
}
```
