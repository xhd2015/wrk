# Scenario

**Feature**: `--list` + `--exec` is rejected

```
myrepo -> wrk --list --exec true -> non-zero; not valid / mutually exclusive
```

## Steps

1. Initialize git repo.
2. Run `wrk --list --exec true`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--list", "--exec", "true"}
	return nil
}
```
