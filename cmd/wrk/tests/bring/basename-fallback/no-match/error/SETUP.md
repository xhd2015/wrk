# Scenario

**Feature**: wrk --bring mydep with no match reports does not exist

```
consumer/ (cwd, no ./mydep) -> wrk --bring mydep -> wrk: <abs>/mydep does not exist
```

## Steps

1. Create consumer git repo with go.mod requiring `example.com/dep`.
2. Leave `projects.json` without a matching `mydep` entry.
3. Run `wrk --bring mydep` from consumer cwd.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerForBringBasename(t, req.WorkRoot)
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", "mydep"}
	return nil
}
```