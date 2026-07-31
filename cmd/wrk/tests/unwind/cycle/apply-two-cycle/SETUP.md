# Scenario

**Feature**: apply-mode (non-dry-run) still rejects cycle before any mutation

```
# cycle-a requires cycle-b; cycle-b requires cycle-a; both nested dirty
A ↔ B -> wrk --unwind --tag-next --push --done   # NO --dry-run
  -> Error: cycle …; exit ≠ 0
  -> no peel apply; HEADs / worktrees unchanged
```

## Steps

1. Build host root linked wt with external cycle-a and cycle-b mutual requires.
2. Dirtify both cycle members.
3. Run **apply-mode** unwind (no `--dry-run`) with full ship/land flags.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupTwoCycleStack(t, req)
	// Flags would be valid if acyclic; cycle still aborts first (before apply stub).
	req.Args = []string{"--unwind", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
