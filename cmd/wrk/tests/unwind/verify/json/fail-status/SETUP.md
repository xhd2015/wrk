# Scenario

**Feature**: JSON reports check status fail and summary.result fail on residual drift

```
# require-drift fixture -> wrk --unwind --verify --json
  -> require-drift status fail; summary.result fail; exit 1
  -> pure JSON; no Error: for logical FAIL
```

## Steps

1. Seed require-drift fixture.
2. Run verify with `--json`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifyRequireDrift(t, req)
	req.Args = verifyJSONArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
