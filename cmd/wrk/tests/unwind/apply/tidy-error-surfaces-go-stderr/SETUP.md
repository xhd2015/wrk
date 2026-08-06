# Scenario

**Bug**: `go mod tidy` failure discards go child stderr (only `exit status 1`)

```
# single-module leaf pin stack, but modproxy has only v0.0.1 (no next)
root main + leaf ext (dirty+ahead)
  -> wrk --unwind --done --tag-next --push
  -> peel leaf → tag v0.0.2 → pin consumer require → v0.0.2
  -> go mod tidy fails resolving next (missing proxy entry)
  -> error must include go child diagnostic body, not only exit status 1
```

## Steps

1. Build tidy-error pin stack (`setupApplyTidyErrorPinStack`) — peel+pin reach
   tidy; next version intentionally absent from file proxy.
2. Run non-dry-run unwind with land + pin flags from root main.

## Context

- Expect **RED** while `goModTidy` discards child streams without `-v`.
- GREEN when tidy failure wraps/returns trimmed go stderr (e.g. `no such file`,
  `unknown revision`, `reading …@v/….zip`).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyTidyErrorPinStack(t, req)
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
