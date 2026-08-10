# Scenario

**Feature**: tidy fails mid partial-edit → restore WIP exactly (P3-4)

```
# multi-repo free-first; root go.mod dirty WIP; modproxy has only v0.0.1 (no next)
leaf → pin root require next → go mod tidy fails on Base
  -> wrk --unwind --tag-next --push --done
  -> non-zero; WT go.mod restored to pre-run WIP snapshot (byte-identical)
  -> no half-mutated Base left on worktree
```

## Steps

1. Seed multi-repo cascade stack with old-only modproxy + dirty go.mod WIP
   (`setupApplyCascadePartialEditTidyFail` snapshots WIP).
2. Run land + pin flags without `--add-all`.
3. Expect non-zero; exact WIP restore.

## Context

- Fail path: restore from save before return; do not leave Base-only or
  half-pinned content on WT.
- Leaf may already be tagged before consumer tidy fails (fail-fast at pin/tidy).
- **RED** until partial-edit restore-on-error lands (today may hard-fail before
  any save, or leave mutations — assert exact restore).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadePartialEditTidyFail(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push", "--done"}
	return nil
}
```
