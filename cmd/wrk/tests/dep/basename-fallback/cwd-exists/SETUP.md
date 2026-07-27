# Scenario

**Feature**: local ./basename in consumer cwd blocks --dep basename fallback

```
# ./<basename> exists under consumer cwd (even non-git)
saved/<basename> recorded -> wrk --dep <basename> -> use local path, no projects.json lookup
```

## Steps

- Descendants create a local non-git `./<basename>` directory inside the consumer repo.
- A saved dep with the same basename is also recorded in `projects.json`.
- Run `wrk --dep <basename>` from consumer cwd.

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