# Scenario

**Feature**: wrk --dep errors when dep module is not in consumer go.mod

```
# consumer without dep require -> wrk --dep -> non-zero
```

## Steps

1. Create consumer **without** requiring dep.
2. Create valid dep repo.
3. Run `wrk --dep`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerRepo(t, req.WorkRoot, false)
	depPath := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = depPath
	req.Args = []string{"--dep", depPath}
	return nil
}
```