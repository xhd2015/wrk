# Scenario

**Feature**: `wrk --status` from linked consumer lists consumer + two external deps

```
consumerWt + external/aaa-dep + external/zzz-dep
  + incomplete warm index
  -> wrk --status
  -> 3 Dir blocks; both external paths present
```

## Steps

1. Parent builds two-external fixture + incomplete warm seed.
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
