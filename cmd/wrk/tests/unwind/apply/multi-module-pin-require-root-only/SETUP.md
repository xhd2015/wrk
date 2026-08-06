# Scenario

**Bug**: multi-module dep peel force-adds nested module require the consumer never needed

```
# dep multi-module: example.com/dep + example.com/dep/nested; tag root next only
# consumer requires only example.com/dep@v0.0.1 (no nested)
root main + dep ext (dirty+ahead root pkg)
  -> wrk --unwind --done --tag-next --push
  -> peel dep: land → tag v0.0.2 (root) → push
  -> Pin consumer: bump example.com/dep → v0.0.2
  -> must NOT require example.com/dep/nested; go mod tidy OK; exit 0
```

## Steps

1. Build multi-module dep + require-root-only consumer
   (`setupApplyMultiModuleRootOnlyPinStack`).
2. Run non-dry-run unwind with land + pin flags from root main.

## Context

- Expect **RED** while product Cartesian-pins every dep module dir into the
  consumer: **two** `pin root <- dep` log lines (one per module dir). Tidy may
  drop an unused force-added nested require, so pin-line count is the primary
  surface; go.mod must still lack nested after success.
- GREEN when pin matches consumer require/replace only (single pin line + root
  require bumped; nested absent).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyMultiModuleRootOnlyPinStack(t, req)
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
