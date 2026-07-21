# Scenario

**Feature**: --task with description that slugifies to empty produces an error

```
wrk --task "!!!" -> slug is empty -> non-zero exit, error
```

## Steps

1. Create repo.
2. Set req.TaskDesc = "!!!".
3. Verify non-zero exit — slug is empty after sanitization.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.TaskDesc = "!!!"
	return nil
}
```