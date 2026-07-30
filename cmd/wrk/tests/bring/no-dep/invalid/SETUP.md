# Scenario

**Feature**: --no-dep is rejected unless combined with --bring

```
# bare --no-dep or --no-dep with other modes -> non-zero
# stderr: --no-dep is only valid with --bring
wrk --no-dep [other non-host mode]
  -> exit non-zero
  -> only-valid-with message
```

## Steps

- Leaves set `req.Args` with invalid host combinations.
- Prefer `InProcess` L2 early-reject path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves set Args; parent bring/no-dep already ensures snapshot helpers.
	req.InProcess = true
	return nil
}
```
