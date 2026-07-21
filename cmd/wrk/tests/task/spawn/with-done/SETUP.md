# Scenario

**Feature**: --task is mutually exclusive with --done

```
wrk --task "x" --done -> non-zero exit, mutually exclusive
```

## Steps

1. Create repo.
2. Set req.TaskDesc = "x" and req.Args = ["--done"].
3. Verify non-zero exit.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.TaskDesc = "x"
	req.Args = []string{"--done"}
	return nil
}
```