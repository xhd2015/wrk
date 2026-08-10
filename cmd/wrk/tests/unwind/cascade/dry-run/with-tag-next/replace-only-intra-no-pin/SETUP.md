# Scenario

**Feature**: intra-repo replace alone does **not** force cascade pin (D4 control) — C-DR8

```
# dirty root; pkgs/shared clean; require matches; replace => ./pkgs/shared only
root → shared (intra keep-local)
  -> wrk --unwind --dry-run --tag-next
  -> would: peel .
  -> no would: pin … <- … solely for replace
  -> exit 0; zero mutations
```

## Steps

1. Seed single-repo root + `pkgs/shared` with matching require and intra replace.
2. No owned-changed after baseline tags (no tag-next, no require-drift).
3. Dirtify root only; dry-run with `--tag-next`.
4. Expect peel `.` and **no** cascade pin invented for keep-local replace.

## Context

- D4: intra-repo `replace => ./…` alone must not force needs-pin.
- Contrasts C-DR7 (external/droppable replace ⇒ pin) and C-DR1 (owned-changed
  shared still tags then pins).
- Expected **GREEN** on current planner (no pending); stays GREEN after D1 if
  implementer scopes needs-pin to droppable external replaces only.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeReplaceOnlyIntraCleanShared(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next"}
	recordUnwindBaseline(t, req)
	return nil
}
```
