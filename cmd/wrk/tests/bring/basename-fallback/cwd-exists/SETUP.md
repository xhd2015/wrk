# Scenario

**Feature**: local ./basename in consumer cwd blocks --bring basename fallback

```
# ./<basename> exists under consumer cwd (even non-git)
saved/<basename> recorded -> wrk --bring <basename> -> use local path, no projects.json lookup
```

## Steps

- Descendants create a local non-git `./<basename>` directory inside the consumer repo.
- A saved dep with the same basename is also recorded in `projects.json`.
- Run `wrk --bring <basename>` from consumer cwd.

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