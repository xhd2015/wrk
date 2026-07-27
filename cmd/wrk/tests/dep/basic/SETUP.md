# Scenario

**Feature**: wrk --dep creates external worktree, replace, tidy, and gitignore entry

```
# consumer requires dep -> wrk --dep dep-repo -> external/mydep-main-{date} + replace + /external gitignore
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create dep git repo `mydep` with module `example.com/dep`.
3. Run `wrk --dep <dep>` from consumer.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	depPath := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", depPath}
	return nil
}
```