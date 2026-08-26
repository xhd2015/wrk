# Scenario

**Feature**: external dep preferred branch pre-exists → path stays free basename; branch only `-1`

```
# dep main has refs/heads/main-{date}; path external/mydep free
consumer --bring mydep
  -> path external/mydep
  -> branch main-{date}-1 (no mydep- basename; -b new branch)
  -> stderr warning: branch main-{date} exists; using main-{date}-1
```

## Steps

1. Create consumer requiring `example.com/dep`.
2. Create dep repo `mydep` on `main`.
3. Pre-create branch `main-{WRK_DATE}` in the dep repo.
4. Run `wrk --bring <dep>` from consumer.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	depPath := initBringDepRepo(t, req.WorkRoot, "mydep", true)
	runGitIsolated(t, depPath, "branch", branchName("main", wrkDate, 0))

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```
