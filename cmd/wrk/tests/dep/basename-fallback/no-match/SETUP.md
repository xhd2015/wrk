# Scenario

**Feature**: --dep basename with zero saved matches reports does not exist

```
# no local ./basename and no projects.json match
consumer -> wrk --dep <basename> -> wrk: <candidate> does not exist
```

## Steps

- Descendants use consumer cwd without local `./<basename>` and empty or non-matching `projects.json`.
- Run `wrk --dep <basename>`.

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