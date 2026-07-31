# Scenario

**Feature**: two-repo require cycle is rejected on dry-run

```
# cycle-a requires cycle-b; cycle-b requires cycle-a; both nested dirty
A ↔ B -> wrk --unwind --dry-run --tag-next --push --done
  -> Error: cycle …; exit ≠ 0
  -> no multi-step would: peel plan; HEADs unchanged
```

## Steps

1. Build host root linked wt with external `cycle-a` and `cycle-b` mutual requires.
2. Dirtify both cycle members.
3. Run unwind dry-run with full ship/land flags (flags present must not bypass cycle check).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoCycleStack(t, req)
	// Flags would be valid if acyclic; cycle still aborts first.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
