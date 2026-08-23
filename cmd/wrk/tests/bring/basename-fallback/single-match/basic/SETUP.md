# Scenario

**Feature**: wrk --bring mydep creates external worktree from single saved dep

```
saved/mydep in projects.json (module example.com/dep)
consumer requires dep, no ./mydep -> wrk --bring mydep -> external/mydep
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create dep git repo at `{WorkRoot}/saved/mydep` with module `example.com/dep`.
3. Record saved dep with `wrk --add`.
4. Run `wrk --bring mydep` from consumer cwd.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerForBringBasename(t, req.WorkRoot)
	savedDep := initSavedDepRepo(t, req.WorkRoot, "saved", "mydep")
	recordSavedProject(t, req, savedDep)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = savedDep
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", "mydep"}
	return nil
}
```