# Scenario

**Feature**: `wrk --repos` shares status discovery — lists consumer + external dep

```
same fixture as one-external/status
  -> wrk --repos from consumerWt
  -> .\nexternal/mydep-main-{date}\n
```

## Steps

1. Parent builds linked consumer + one external dep + incomplete warm seed.
2. Override Args to `--repos`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--repos"}
	return nil
}
```
