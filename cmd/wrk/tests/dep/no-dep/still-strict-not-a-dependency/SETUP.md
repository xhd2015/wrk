# Scenario

**Feature**: --dep --no-dep still hard-errors when dep is not required (no worktree)

```
# consumer without require -> wrk --dep <dep> --no-dep
#   -> non-zero; same class of "not a dependency" error as plain --dep
#   -> no external worktree (strict analyse first)
consumer (no require) + mydep -> wrk --dep <dep> --no-dep -> error, no wt
```

## Steps

1. Create consumer without requiring dep.
2. Create valid dep repo.
3. Run `wrk --dep <dep> --no-dep`.

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
	req.ConsumerTop = consumer
	req.Args = []string{"--dep", depPath, "--no-dep"}
	return nil
}
```
