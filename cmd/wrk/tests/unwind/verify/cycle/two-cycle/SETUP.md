# Scenario

**Feature**: two-repo require cycle is rejected on verify (no success body)

```
# cycle-a requires cycle-b; cycle-b requires cycle-a; both nested dirty
A ↔ B -> wrk --unwind --verify
  -> Error: cycle …; exit ≠ 0
  -> no verify banners; HEADs unchanged
```

## Steps

1. Build host root linked wt with external cycle-a and cycle-b mutual requires.
2. Dirtify both cycle members (parent helper).
3. Run verify only (no dry-run / apply partners).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoCycleStack(t, req)
	req.Args = verifyArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
