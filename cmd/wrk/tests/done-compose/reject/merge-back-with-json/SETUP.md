# Scenario

**Feature**: `--merge-back --json` is rejected

```
myrepo -> wrk --merge-back --json
  -> non-zero
  -> stderr names --json and --merge-back
```

## Steps

1. `initGitRepoOnMain` under WorkRoot.
2. Run `wrk --merge-back --json` from the main repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--merge-back", "--json"}
	return nil
}
```
