# Scenario

**Feature**: --bring argument with path separator skips basename fallback

```
# <dir> contains '/' or '\' -> no projects.json lookup
saved/<basename> recorded -> wrk --bring sub/<basename> -> does not exist (no fallback)
```

## Steps

- Descendants record a saved dep whose basename would match if fallback ran.
- Run `wrk --bring <path-with-separator>` from consumer cwd without that relative path.

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