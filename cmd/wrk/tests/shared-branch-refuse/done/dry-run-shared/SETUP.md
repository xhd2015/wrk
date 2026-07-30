# Scenario

**Feature**: `wrk --done --dry-run` fail-closed on shared branch (hard error, not plan success)

```
# shared two live checkouts
myrepo + wt1 + wt2 same branch
  -> wrk --done --dry-run
  -> non-zero; Error: refuse --done
  -> no would-plan treated as success; zero mutations
```

## Steps

1. Shared two-linked fixture.
2. Run `wrk --done --dry-run` from primary wt.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedTwoLinked(t, req)
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
