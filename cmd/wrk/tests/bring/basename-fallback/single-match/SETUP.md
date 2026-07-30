# Scenario

**Feature**: single saved dep match resolves --bring basename to saved path

```
# one projects.json entry matches basename; consumer cwd has no local ./basename
consumer -> wrk --bring mydep -> external worktree from saved/mydep
```

## Steps

- Descendants seed exactly one saved dep project whose basename matches `--bring` argument.
- Run `wrk --bring <basename>` from consumer cwd without a local `./<basename>` entry.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBringBasenameFallbackHelpersUsed()
	return nil
}
```