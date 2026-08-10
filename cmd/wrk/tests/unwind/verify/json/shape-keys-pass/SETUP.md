# Scenario

**Feature**: JSON top-level and nested keys for clean pass stack

```
clean tagged main -> wrk --unwind --verify --json
  -> keys: work_dir, checks, summary, warnings
  -> 6 checks all status pass; summary.result pass; exit 0
```

## Steps

1. Seed clean tagged single main.
2. Run verify with `--json`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVerifySingleMainClean(t, req)
	req.Args = verifyJSONArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
