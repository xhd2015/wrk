# Scenario

**Feature**: single saved dep match resolves --dep basename to saved path

```
# one projects.json entry matches basename; consumer cwd has no local ./basename
consumer -> wrk --dep mydep -> external worktree from saved/mydep
```

## Steps

- Descendants seed exactly one saved dep project whose basename matches `--dep` argument.
- Run `wrk --dep <basename>` from consumer cwd without a local `./<basename>` entry.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureDepBasenameFallbackHelpersUsed()
	return nil
}
```