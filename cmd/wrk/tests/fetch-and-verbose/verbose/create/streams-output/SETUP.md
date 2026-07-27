# Scenario

**Feature**: wrk -v create streams git worktree add subprocess output to stderr

```
main repo -> wrk -v -> stderr has timestamp worktree add log + git's own lines
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "create-v-stream")
	initFetchVerboseRepo(t, repo, "create v stream")
	req.RepoDir = repo
	req.Args = []string{"--new", "-v"}
	return nil
}
```