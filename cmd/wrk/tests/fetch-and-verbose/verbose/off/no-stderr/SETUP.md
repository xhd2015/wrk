# Scenario

**Feature**: wrk --list without -v has empty stderr

```
main repo -> wrk --list -> stderr empty
```

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repo := filepath.Join(req.WorkRoot, "off-list-main")
	initFetchVerboseRepo(t, repo, "off list main")
	req.RepoDir = repo
	req.Args = []string{"--list"}
	return nil
}
```