# Scenario

**Bug**: B1 apply — 3-level freeHost mid pinReady pins dirty free @ NextTag **before** cascade tag-next (T-tag1)

```
# production shape: top → mid freeHost → dirty free leaf (spl → kool → go-pkgs)
# free owned-changed → planned next v0.0.2; mid has external replace + FEATURE_WIP
# B1: free + mid freeHost peel early; pure top pinConsumer deferred
leaf free + mid freeHost + top
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free peel/land first
  -> mid early peel: pinReady must SKIP untagged free (wait for cascade)
  -> cascade tag-next free @ v0.0.2 (+ push) then pin mid <- free @ v0.0.2
  -> mid require @ v0.0.2; no external replace; exit 0
```

## Steps

1. Seed T-tag1 fixture (`setupApplyPinBeforeFeatureThreeLevelFreeHostDirty`):
   dirty free leaf + mid freeHost (replace + feature WIP) + top pure pinConsumer
   + modproxy old/next + bare origins.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert free tag-before-mid-pin order, free advanced/tagged, mid require@next.

## Context

- Encodes production failure: mid is **pinConsumer of free** and **freeHost of top**
  → peels early; `pinReadyExternalReplacesBeforeGenCommit` runs before global
  cascade tag-next. Without `attachTagScopeToModules`, untagged NextTag looks
  "ready" → premature pin (prod: `unknown revision …/v0.0.108`).
- L2 locks **tag-next free before pin mid←free @ next** (T-tag2 folded here).
- No no-local-replace hook on mid: after fix, replace may remain through mid
  feature gen-commit until cascade pin (dirty free is not pinReady-ready).
- Sealed T1/T2/T-M1 ASSERT meaning unchanged.
- Classic TDD **RED** until pinReady attaches tagscope / refuses untagged next.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureThreeLevelFreeHostDirty(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```
