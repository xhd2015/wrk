# Scenario

**Feature**: wrk --new -v create logs worktree add

```
main repo -> wrk --new -v -> stderr contains worktree add
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "create-v-main")
	initFetchVerboseRepo(t, repo, "create v main")
	req.RepoDir = repo
	req.Args = []string{"--new", "-v"}
	return nil
}
```