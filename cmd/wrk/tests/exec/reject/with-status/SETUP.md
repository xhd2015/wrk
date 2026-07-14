# Scenario

**Feature**: `--status` + `--exec` is rejected

```
myrepo -> wrk --status --exec true -> non-zero; not valid / mutually exclusive
```

## Steps

1. Initialize git repo.
2. Run `wrk --status --exec true`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--status", "--exec", "true"}
	return nil
}
```
