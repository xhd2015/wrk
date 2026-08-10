# Scenario

**Feature**: C1 free monorepo multi-module tags (root + cmd) + consumer pins root only

```
# free go-pkgs shape: example.com/dot-pkgs + example.com/dot-pkgs/cmd (both owned-changed)
# consumer requires free root only + droppable external replace (never free/cmd)
free multi-module ext + consumer root-only
  -> wrk --unwind --done --tag-next --push
  -> peel free: land → free-first cascade on free monorepo
  -> tag free root @ v0.0.2; pin free/cmd ← free; tag free/cmd @ cmd/v0.0.2
  -> pin consumer ← free root @ v0.0.2 only (no nested require); exit 0
```

## Steps

1. Seed C1 fixture (`setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly`):
   free monorepo root+cmd both next-tag; consumer require root only + replace.
2. Run apply with land + `--tag-next --push` (no gen-commit — pin selectivity focus).
3. Assert free multi-tag free-first order, consumer root-only pin, no nested force-add.

## Context

- Distinct from `apply/multi-module-pin-require-root-only` (nested clean / no cascade
  free/cmd pin) and from T2 (single-module free).
- Planner must tag free root before pin of nested free/cmd when free-first orders
  that way; consumer pin pairs only modules it requires/replaces.
- P3 coverage backfill: may be GREEN if multi-module pin selectivity already
  covers both-modules-owned-changed free monorepo; RED if cartesian pin or wrong
  nested version.
- Do not rewrite sealed T1–T-spl / P2 ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly(t, req)
	// Clean multi-repo apply path: land free, cascade tags+pins, push.
	// No gen-commit — C1 locks multi-module free tags + consumer root-only pin.
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
