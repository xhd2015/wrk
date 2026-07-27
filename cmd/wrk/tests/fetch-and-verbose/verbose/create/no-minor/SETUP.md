# Scenario

**Feature**: wrk -v create does not log minor git reads

```
stderr has no rev-parse or status log lines
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "create-v-nominor")
	initFetchVerboseRepo(t, repo, "create v nominor")
	req.RepoDir = repo
	req.Args = []string{"--new", "-v"}
	return nil
}
```