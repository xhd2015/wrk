# Scenario

**Feature**: external stack replace alone plans cascade pin (clean free dep, no drift) — C-DR7

```
# root dirty; leaf external clean at v0.0.1; require matches; replace => external/…
root ← leaf (external replace only; no owned-changed / no require-drift)
  -> wrk --unwind --dry-run --tag-next
  -> would: peel .
  -> would: pin example.com/root <- example.com/dot-pkgs @ v0.0.1
  -> no would: peel external/…; no would: tag-next leaf
  -> exit 0; zero mutations
```

## Steps

1. Seed multi-repo stack: dirty root + clean nested leaf external at tag `v0.0.1`.
2. Root `go.mod`: require leaf@v0.0.1 (matches) **and** `replace => ./external/…`.
3. Run dry-run with `--tag-next` only (no residual dirty edges → no `--push`/`--done`).
4. Expect peel `.` then cascade pin at **current require** version; no leaf peel / tag-next.

## Context

- Planner today only pending-marks from owned-changed / require-drift / fixpoint —
  **replace-only never plans pin** → Classic **RED** until D1 lands.
- D3: when no tag/drift, pin version = keep current require (`v0.0.1`).
- D4 control is the sibling leaf `replace-only-intra-no-pin` (intra replace alone).
- Peels-then-cascade order stays sealed (`assertCascadeAfterPeels`).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeReplaceOnlyExternalCleanDep(t, req)
	// Only root dirty (main) → no NeedsLand / HasPendingEdges; --tag-next enables cascade.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next"}
	recordUnwindBaseline(t, req)
	return nil
}
```
