# Scenario

**Feature**: `wrk --status` from linked consumer lists consumer + external dep worktree

```
consumerWt + external/mydep-main-{date}
  + incomplete warm index
  -> wrk --status
  -> Dir: .  and  Dir: external/mydep-main-{date}
```

## Steps

1. Parent builds linked consumer + one external dep + incomplete warm seed.
2. Run `wrk --status` from consumer (InProcess).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--status"}
	return nil
}
```
