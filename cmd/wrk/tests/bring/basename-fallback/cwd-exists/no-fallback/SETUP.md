# Scenario

**Feature**: consumer ./mydep blocks fallback even when saved dep exists

```
consumer/mydep (non-git dir exists)
saved/mydep recorded in projects.json
consumer cwd -> wrk --bring mydep -> not a git repository (no fallback to saved)
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Create non-git directory `{consumer}/mydep`.
3. Create and record saved dep at `{WorkRoot}/saved/mydep`.
4. Run `wrk --bring mydep` from consumer cwd.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerForBringBasename(t, req.WorkRoot)
	initLocalNonGitBasenameInDir(t, consumer, "mydep")
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