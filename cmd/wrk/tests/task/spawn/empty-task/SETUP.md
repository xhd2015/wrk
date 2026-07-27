# Scenario

**Feature**: --task with empty description produces an error

```
wrk --task "" -> non-zero exit, stderr says task description must not be empty
```

## Steps

1. Create repo.
2. Use req.Args = ["--task", ""] to pass the empty value.
3. Verify non-zero exit and error message.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--task", ""}
	return nil
}
```