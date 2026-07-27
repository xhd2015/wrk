# Scenario

**Feature**: wrk --dep errors when dep path is git but not a Go module

```
# git repo without go.mod -> wrk --dep -> non-zero
```

## Steps

1. Create consumer with dep require.
2. Create git repo without `go.mod`.
3. Run `wrk --dep`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	depPath := initDepRepo(t, req.WorkRoot, "mydep", false)

	req.RepoDir = consumer
	req.DepPath = depPath
	req.Args = []string{"--dep", depPath}
	return nil
}
```