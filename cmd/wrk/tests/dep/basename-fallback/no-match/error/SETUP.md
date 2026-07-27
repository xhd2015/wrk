# Scenario

**Feature**: wrk --dep mydep with no match reports does not exist

```
consumer/ (cwd, no ./mydep) -> wrk --dep mydep -> wrk: <abs>/mydep does not exist
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Leave `projects.json` without a matching `mydep` entry.
3. Run `wrk --dep mydep` from consumer cwd.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerForDepBasename(t, req.WorkRoot)
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", "mydep"}
	return nil
}
```